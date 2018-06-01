package pdp

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"

	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
	"github.com/infobloxopen/go-trees/uintX/domaintree16"
	"github.com/infobloxopen/go-trees/uintX/domaintree32"
	"github.com/infobloxopen/go-trees/uintX/domaintree64"
	"github.com/infobloxopen/go-trees/uintX/domaintree8"
	"github.com/infobloxopen/go-trees/uintX/iptree16"
	"github.com/infobloxopen/go-trees/uintX/iptree32"
	"github.com/infobloxopen/go-trees/uintX/iptree64"
	"github.com/infobloxopen/go-trees/uintX/iptree8"
	"github.com/infobloxopen/go-trees/uintX/strtree16"
	"github.com/infobloxopen/go-trees/uintX/strtree32"
	"github.com/infobloxopen/go-trees/uintX/strtree64"
	"github.com/infobloxopen/go-trees/uintX/strtree8"
)

// ContentKeyTypes gathers all types which can be a key for content map.
var ContentKeyTypes = makeTypeSet(
	TypeString,
	TypeAddress,
	TypeNetwork,
	TypeDomain,
)

// LocalContentStorage is a storage of all independent local contents.
type LocalContentStorage struct {
	r *strtree.Tree
}

// NewLocalContentStorage creates new LocalContentStorage instance. It's filled
// with given contents.
func NewLocalContentStorage(items []*LocalContent) *LocalContentStorage {
	s := &LocalContentStorage{r: strtree.NewTree()}

	for _, item := range items {
		s.r.InplaceInsert(item.id, item)
	}

	return s
}

// Get returns content item by given content id and nested content item id.
func (s *LocalContentStorage) Get(cID, iID string) (*ContentItem, error) {
	v, ok := s.r.Get(cID)
	if !ok {
		return nil, newMissingContentError(cID)
	}

	cnt, ok := v.(*LocalContent)
	if !ok {
		return nil, newInvalidContentStorageItem(cID, v)
	}

	item, err := cnt.Get(iID)
	if err != nil {
		return nil, bindError(err, cID)
	}

	return item, nil
}

// Add puts new content to storage. It returns copy of existing storage with
// new content in it. Existing storage isn't affected by the operation.
func (s *LocalContentStorage) Add(c *LocalContent) *LocalContentStorage {
	return &LocalContentStorage{r: s.r.Insert(c.id, c)}
}

// GetLocalContent returns content from storage by given id only if the content
// has its own tag and the tag matches to tag argument.
func (s *LocalContentStorage) GetLocalContent(cID string, tag *uuid.UUID) (*LocalContent, error) {
	v, ok := s.r.Get(cID)
	if !ok {
		return nil, newMissingContentError(cID)
	}

	c, ok := v.(*LocalContent)
	if !ok {
		return nil, newInvalidContentStorageItem(cID, v)
	}

	if c.tag == nil {
		return nil, newUntaggedContentModificationError(cID)
	}

	if tag == nil {
		return nil, newMissingContentTagError()
	}

	if c.tag.String() != tag.String() {
		return nil, newContentTagsNotMatchError(cID, c.tag, tag)
	}

	return c, nil
}

// NewTransaction creates new transaction for given content in the storage.
func (s *LocalContentStorage) NewTransaction(cID string, tag *uuid.UUID) (*LocalContentStorageTransaction, error) {
	c, err := s.GetLocalContent(cID, tag)
	if err != nil {
		return nil, err
	}

	return &LocalContentStorageTransaction{
		tag:     *tag,
		ID:      cID,
		items:   c.items,
		symbols: c.symbols.makeROCopy(),
	}, nil
}

// String implements Stringer interface.
func (s *LocalContentStorage) String() string {
	if s == nil {
		return ""
	}

	lines := []string{"content:"}
	for p := range s.r.Enumerate() {
		line := fmt.Sprintf("- %s: ", p.Key)
		if c, ok := p.Value.(*LocalContent); ok {
			line += c.String()
		} else {
			line += fmt.Sprintf("\"invalid type: %T\"", p.Value)
		}

		lines = append(lines, line)
	}

	if len(lines) > 1 {
		return strings.Join(lines, "\n")
	}

	return ""
}

// ContentUpdate encapsulates list of changes to particular content.
type ContentUpdate struct {
	cID    string
	oldTag uuid.UUID
	newTag uuid.UUID
	cmds   []*command
}

// NewContentUpdate creates empty update for given content and sets tags.
// Content must have oldTag so update can be applied. newTag will be set
// to content after the update.
func NewContentUpdate(cID string, oldTag, newTag uuid.UUID) *ContentUpdate {
	return &ContentUpdate{
		cID:    cID,
		oldTag: oldTag,
		newTag: newTag,
		cmds:   []*command{}}
}

// Append inserts particular change to the end of changes list. Op is an
// operation (like add or delete), path identifies content part to perform
// operation and entity item to add (and ignored in case of delete operation).
func (u *ContentUpdate) Append(op int, path []string, entity *ContentItem) {
	u.cmds = append(u.cmds, &command{op: op, path: path, entity: entity})
}

// String implements Stringer interface.
func (u *ContentUpdate) String() string {
	if u == nil {
		return "no content update"
	}

	lines := []string{fmt.Sprintf("content update: %s - %s\ncontent: %q", u.oldTag, u.newTag, u.cID)}
	if len(u.cmds) > 0 {
		lines = append(lines, "commands:")
		for _, cmd := range u.cmds {
			lines = append(lines, "- "+cmd.describe())
		}
	}

	return strings.Join(lines, "\n")
}

// LocalContentStorageTransaction represents transaction for local content.
// Transaction aggregates updates and then can be committed
// to LocalContentStorage to make all the updates visible at once.
type LocalContentStorageTransaction struct {
	tag     uuid.UUID
	ID      string
	items   *strtree.Tree
	symbols Symbols
	err     error
}

// Symbols returns symbol tables captured from content storage on transaction
// creation.
func (t *LocalContentStorageTransaction) Symbols() Symbols {
	return t.symbols
}

func (t *LocalContentStorageTransaction) applyCmd(cmd *command) error {
	switch cmd.op {
	case UOAdd:
		return t.add(cmd.path, cmd.entity)

	case UODelete:
		return t.del(cmd.path)
	}

	return newUnknownContentUpdateOperationError(cmd.op)
}

// Apply updates captured content with given content update.
func (t *LocalContentStorageTransaction) Apply(u *ContentUpdate) error {
	if t.err != nil {
		return newFailedContentTransactionError(t.ID, t.tag, t.err)
	}

	if t.ID != u.cID {
		return newContentTransactionIDNotMatchError(t.ID, u.cID)
	}

	if t.tag.String() != u.oldTag.String() {
		return newContentTransactionTagsNotMatchError(t.ID, t.tag, u.oldTag)
	}

	for i, cmd := range u.cmds {
		err := t.applyCmd(cmd)
		if err != nil {
			t.err = err
			return bindErrorf(err, "command %d - %s", i, cmd.describe())
		}
	}

	t.tag = u.newTag
	return nil
}

// Commit creates new content storage with updated content previously captured
// by transaction. Each commit creates copy of storage with only its changes
// applied. So applications must ensure that all commits to the same storage
// are made sequentially and that there is only one transaction for the same
// content id (all pairs of NewTransaction and Commit for the same content id
// go sequentially).
func (t *LocalContentStorageTransaction) Commit(s *LocalContentStorage) (*LocalContentStorage, error) {
	if t.err != nil {
		return nil, newFailedContentTransactionError(t.ID, t.tag, t.err)
	}

	c := &LocalContent{id: t.ID, tag: &t.tag, items: t.items}
	if s == nil {
		return NewLocalContentStorage([]*LocalContent{c}), nil
	}

	return s.Add(c), nil
}

func (t *LocalContentStorageTransaction) getItem(ID string) (*ContentItem, error) {
	v, ok := t.items.Get(ID)
	if !ok {
		return nil, newMissingContentItemError(ID)
	}

	c, ok := v.(*ContentItem)
	if !ok {
		return nil, bindError(newInvalidContentItemError(v), ID)
	}

	return c, nil
}

func (t *LocalContentStorageTransaction) parsePath(c *ContentItem, rawPath []string) ([]AttributeValue, error) {
	if len(rawPath) > len(c.k) {
		return nil, newTooLongRawPathContentModificationError(c.k, rawPath)
	}

	path := make([]AttributeValue, len(rawPath))
	for i, s := range rawPath {
		if c.k[i] == TypeAddress || c.k[i] == TypeNetwork {
			a := net.ParseIP(s)
			if a != nil {
				path[i] = MakeAddressValue(a)
				continue
			}

			_, n, err := net.ParseCIDR(s)
			if err == nil {
				path[i] = MakeNetworkValue(n)
				continue
			}

			return nil, bindErrorf(newInvalidAddressNetworkStringCastError(s, err), "%d", i+2)
		}

		v, err := MakeValueFromString(c.k[i], s)
		if err != nil {
			return nil, bindErrorf(err, "%d", i+2)
		}

		path[i] = v
	}

	return path, nil
}

func (t *LocalContentStorageTransaction) add(rawPath []string, v interface{}) error {
	if len(rawPath) <= 0 {
		return bindError(newTooShortRawPathContentModificationError(), t.ID)
	}

	ID := rawPath[0]

	if len(rawPath) > 1 {
		c, err := t.getItem(ID)
		if err != nil {
			return bindError(err, t.ID)
		}

		path, err := t.parsePath(c, rawPath[1:])
		if err != nil {
			return bindError(err, t.ID)
		}

		c, err = c.add(ID, path, v)
		if err != nil {
			return bindError(bindError(err, ID), t.ID)
		}

		t.items = t.items.Insert(ID, c)
		return nil
	}

	c, ok := v.(*ContentItem)
	if !ok {
		return bindError(bindError(newInvalidContentItemError(v), ID), t.ID)
	}

	c.id = ID

	t.items = t.items.Insert(ID, c)
	return nil
}

func (t *LocalContentStorageTransaction) del(rawPath []string) error {
	if len(rawPath) <= 0 {
		return bindError(newTooShortRawPathContentModificationError(), t.ID)
	}

	ID := rawPath[0]

	if len(rawPath) > 1 {
		c, err := t.getItem(ID)
		if err != nil {
			return bindError(err, t.ID)
		}

		path, err := t.parsePath(c, rawPath[1:])
		if err != nil {
			return bindError(err, t.ID)
		}

		c, err = c.del(ID, path)
		if err != nil {
			return bindError(bindError(err, ID), t.ID)
		}

		t.items = t.items.Insert(ID, c)
		return nil
	}

	items, ok := t.items.Delete(ID)
	if !ok {
		return bindError(newMissingContentItemError(ID), t.ID)
	}

	t.items = items
	return nil
}

// LocalContent represents content object which can be accessed by its id and
// independently taged and updated. It holds content items which represent
// mapping objects (or immediate values) of different type.
type LocalContent struct {
	id      string
	tag     *uuid.UUID
	items   *strtree.Tree
	symbols Symbols
}

// NewLocalContent creates content of given id with given tag and set of content
// items. Nil tag makes the content untagged. Such content can't be
// incrementally updated.
func NewLocalContent(id string, tag *uuid.UUID, symbols Symbols, items []*ContentItem) *LocalContent {
	c := &LocalContent{
		id:      id,
		tag:     tag,
		items:   strtree.NewTree(),
		symbols: symbols,
	}

	for _, item := range items {
		c.items.InplaceInsert(item.id, item)
	}

	return c
}

// Get returns content item of given id.
func (c *LocalContent) Get(ID string) (*ContentItem, error) {
	v, ok := c.items.Get(ID)
	if !ok {
		return nil, newMissingContentItemError(ID)
	}

	item, ok := v.(*ContentItem)
	if !ok {
		return nil, bindError(newInvalidContentItemError(v), ID)
	}

	return item, nil
}

// String implements Stringer interface.
func (c *LocalContent) String() string {
	if c == nil {
		return "null"
	}

	if c.tag == nil {
		return "no tag"
	}

	return c.tag.String()
}

// ContentItem represents item of particular content. It can be mapping object
// with defined set of keys to access value of particular type or immediate
// value of defined type.
type ContentItem struct {
	id string
	r  ContentSubItem
	t  Type
	k  []Type
}

// MakeContentValueItem creates content item which represents immediate value
// of given type.
func MakeContentValueItem(id string, t Type, v interface{}) *ContentItem {
	return &ContentItem{
		id: id,
		r:  MakeContentValue(v),
		t:  t}
}

// MakeContentMappingItem creates mapping content item. Argument t is type
// of final value while k list is a list of types from ContentKeyTypes and
// defines which maps the item consists from.
func MakeContentMappingItem(id string, t Type, k []Type, v ContentSubItem) *ContentItem {
	return &ContentItem{
		id: id,
		r:  v,
		t:  t,
		k:  k}
}

// GetType returns content item type
func (c *ContentItem) GetType() Type {
	return c.t
}

func (c *ContentItem) typeCheck(path []AttributeValue, v interface{}) (ContentSubItem, error) {
	item, ok := v.(*ContentItem)
	if !ok {
		return nil, newInvalidContentUpdateDataError(v)
	}

	if item.t != c.t {
		return nil, newInvalidContentUpdateResultTypeError(item.t, c.t)
	}

	if len(path) < len(c.k) {
		if len(path)+len(item.k) != len(c.k) {
			return nil, newInvalidContentUpdateKeysError(len(path), item.k, c.k)
		}

		for i, k := range item.k {
			if k != c.k[len(path)+i] {
				return nil, newInvalidContentUpdateKeysError(len(path), item.k, c.k)
			}
		}

		switch c.k[len(path)] {
		default:
			return nil, newInvalidContentKeyTypeError(c.k[len(path)], ContentKeyTypes)

		case TypeString:
			if _, ok := item.r.(ContentStringMap); !ok {
				return nil, newInvalidContentStringMapError(v)
			}

		case TypeAddress, TypeNetwork:
			if _, ok := item.r.(ContentNetworkMap); !ok {
				return nil, newInvalidContentNetworkMapError(v)
			}

		case TypeDomain:
			if _, ok := item.r.(ContentDomainMap); !ok {
				return nil, newInvalidContentDomainMapError(v)
			}
		}

		return item.r, nil
	}

	subItem, ok := item.r.(ContentValue)
	if !ok {
		return nil, newInvalidContentValueError(v)
	}

	if t, ok := c.t.(*FlagsType); ok {
		switch t.Capacity() {
		case 8:
			if _, ok := subItem.value.(uint8); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case 16:
			if _, ok := subItem.value.(uint16); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case 32:
			if _, ok := subItem.value.(uint32); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case 64:
			if _, ok := subItem.value.(uint64); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}
		}
	} else {
		switch c.t {
		default:
			return nil, newUnknownContentItemResultTypeError(c.t)

		case TypeUndefined:
			return nil, newInvalidContentItemResultTypeError(c.t)

		case TypeBoolean:
			if _, ok := subItem.value.(bool); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeInteger:
			if _, ok := subItem.value.(int64); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeFloat:
			if _, ok := subItem.value.(float64); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeString:
			if _, ok := subItem.value.(string); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeAddress:
			if _, ok := subItem.value.(net.IP); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeNetwork:
			if _, ok := subItem.value.(*net.IPNet); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeDomain:
			if _, ok := subItem.value.(string); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeSetOfStrings:
			if _, ok := subItem.value.(*strtree.Tree); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeSetOfNetworks:
			if _, ok := subItem.value.(*iptree.Tree); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeSetOfDomains:
			if _, ok := subItem.value.(*domaintree.Node); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}

		case TypeListOfStrings:
			if _, ok := subItem.value.([]string); !ok {
				return nil, newInvalidContentValueTypeError(subItem.value, c.t)
			}
		}
	}

	return subItem, nil
}

func (c *ContentItem) add(ID string, path []AttributeValue, v interface{}) (*ContentItem, error) {
	if len(c.k) <= 0 {
		return c, newInvalidContentModificationError()
	}

	if len(path) <= 0 {
		return c, newMissingPathContentModificationError()
	}

	if len(path) > len(c.k) {
		return c, newTooLongPathContentModificationError(c.k, path)
	}

	var err error
	m := c.r

	last := len(path) - 1
	branch := make([]ContentSubItem, last)

	loc := []string{""}

	for i, k := range path[:last] {
		branch[i] = m
		loc = append(loc, k.describe())

		m, err = m.next(k)
		if err != nil {
			return c, bindError(err, strings.Join(loc, "/"))
		}
	}

	k := path[last]
	loc = append(loc, k.describe())

	subItem, err := c.typeCheck(path, v)
	if err != nil {
		return c, bindError(err, strings.Join(loc, "/"))
	}

	m, err = m.put(k, subItem)
	if err != nil {
		return c, bindError(err, strings.Join(loc, "/"))
	}

	for i := len(branch) - 1; i >= 0; i-- {
		p := branch[i]
		m, err = p.put(path[i], m)
		if err != nil {
			return c, bindError(err, strings.Join(loc[:i], "/"))
		}
	}

	return MakeContentMappingItem(ID, c.t, c.k, m), nil
}

func (c *ContentItem) del(ID string, path []AttributeValue) (*ContentItem, error) {
	if len(c.k) <= 0 {
		return c, newInvalidContentModificationError()
	}

	if len(path) <= 0 {
		return c, newMissingPathContentModificationError()
	}

	if len(path) > len(c.k) {
		return c, newTooLongPathContentModificationError(c.k, path)
	}

	var err error
	m := c.r

	last := len(path) - 1
	branch := make([]ContentSubItem, last)

	loc := []string{""}

	for i, k := range path[:last] {
		branch[i] = m
		loc = append(loc, k.describe())

		m, err = m.next(k)
		if err != nil {
			return c, bindError(err, strings.Join(loc, "/"))
		}
	}

	k := path[last]
	loc = append(loc, k.describe())
	m, err = m.del(k)
	if err != nil {
		return c, bindError(err, strings.Join(loc, "/"))
	}

	for i := len(branch) - 1; i >= 0; i-- {
		p := branch[i]
		m, err = p.put(path[i], m)
		if err != nil {
			return c, bindError(err, strings.Join(loc[:i], "/"))
		}
	}

	return MakeContentMappingItem(ID, c.t, c.k, m), nil
}

// Get returns value from content item by given path. It sequentially evaluates
// path expressions and extracts next subitem until gets final value or error.
func (c *ContentItem) Get(path []Expression, ctx *Context) (AttributeValue, error) {
	d := len(path)
	if d != len(c.k) {
		return UndefinedValue, newInvalidSelectorPathError(c.k, path)
	}

	if d > 0 {
		m := c.r
		loc := []string{""}
		for _, e := range path[:d-1] {
			key, err := e.Calculate(ctx)
			if err != nil {
				return UndefinedValue, bindError(err, strings.Join(loc, "/"))
			}

			loc = append(loc, key.describe())

			m, err = m.next(key)
			if err != nil {
				return UndefinedValue, bindError(err, strings.Join(loc, "/"))
			}
		}

		key, err := path[d-1].Calculate(ctx)
		if err != nil {
			return UndefinedValue, bindError(err, strings.Join(loc, "/"))
		}

		v, err := m.getValue(key, c.t)
		if err != nil {
			return UndefinedValue, bindError(err, strings.Join(append(loc, key.describe()), "/"))
		}

		return v, nil
	}

	return c.r.getValue(UndefinedValue, c.t)
}

// ContentSubItem interface abstracts all possible mapping objects and immediate
// content value.
type ContentSubItem interface {
	getValue(key AttributeValue, t Type) (AttributeValue, error)
	next(key AttributeValue) (ContentSubItem, error)
	put(key AttributeValue, v ContentSubItem) (ContentSubItem, error)
	del(key AttributeValue) (ContentSubItem, error)
}

// ContentStringMap implements ContentSubItem as map of string
// to ContentSubItem.
type ContentStringMap struct {
	tree *strtree.Tree
}

// MakeContentStringMap creates instance of ContentStringMap based on strtree
// from github.com/infobloxopen/go-trees. Nodes should be of the same
// ContentSubItem compatible type.
func MakeContentStringMap(tree *strtree.Tree) ContentStringMap {
	return ContentStringMap{tree: tree}
}

func (m ContentStringMap) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	s, err := key.str()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeContentValue(v).getValue(UndefinedValue, t)
}

func (m ContentStringMap) next(key AttributeValue) (ContentSubItem, error) {
	s, err := key.str()
	if err != nil {
		return nil, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return nil, newMissingValueError()
	}

	item, ok := v.(ContentSubItem)
	if !ok {
		return nil, newMapContentSubitemError()
	}

	return item, nil
}

func (m ContentStringMap) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		return MakeContentStringMap(m.tree.Insert(k, v.value)), nil
	}

	return MakeContentStringMap(m.tree.Insert(k, value)), nil
}

func (m ContentStringMap) del(key AttributeValue) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(k)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentStringMap(t), nil
}

// ContentStringFlags8Map implements ContentSubItem as map of string
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 8 bits flags.
type ContentStringFlags8Map struct {
	tree *strtree8.Tree
}

// MakeContentStringFlags8Map creates instance of ContentStringFlags8Map
// based on strtree8 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 8 bits flags.
func MakeContentStringFlags8Map(tree *strtree8.Tree) ContentStringFlags8Map {
	return ContentStringFlags8Map{tree: tree}
}

func (m ContentStringFlags8Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	s, err := key.str()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeFlagsValue8(v, t), nil
}

func (m ContentStringFlags8Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentStringFlags8Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		if n, ok := v.value.(uint8); ok {
			return MakeContentStringFlags8Map(m.tree.Insert(k, n)), nil
		}

		return nil, newInvalidContentStringFlags8MapValueError(v)
	}

	return nil, newInvalidContentValueError(value)
}

func (m ContentStringFlags8Map) del(key AttributeValue) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(k)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentStringFlags8Map(t), nil
}

// ContentStringFlags16Map implements ContentSubItem as map of string
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 16 bits flags.
type ContentStringFlags16Map struct {
	tree *strtree16.Tree
}

// MakeContentStringFlags16Map creates instance of ContentStringFlags16Map
// based on strtree16 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 16 bits flags.
func MakeContentStringFlags16Map(tree *strtree16.Tree) ContentStringFlags16Map {
	return ContentStringFlags16Map{tree: tree}
}

func (m ContentStringFlags16Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	s, err := key.str()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeFlagsValue16(v, t), nil
}

func (m ContentStringFlags16Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentStringFlags16Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		if n, ok := v.value.(uint16); ok {
			return MakeContentStringFlags16Map(m.tree.Insert(k, n)), nil
		}

		return nil, newInvalidContentStringFlags16MapValueError(v)
	}

	return nil, newInvalidContentValueError(value)
}

func (m ContentStringFlags16Map) del(key AttributeValue) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(k)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentStringFlags16Map(t), nil
}

// ContentStringFlags32Map implements ContentSubItem as map of string
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 32 bits flags.
type ContentStringFlags32Map struct {
	tree *strtree32.Tree
}

// MakeContentStringFlags32Map creates instance of ContentStringFlags32Map
// based on strtree32 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 32 bits flags.
func MakeContentStringFlags32Map(tree *strtree32.Tree) ContentStringFlags32Map {
	return ContentStringFlags32Map{tree: tree}
}

func (m ContentStringFlags32Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	s, err := key.str()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeFlagsValue32(v, t), nil
}

func (m ContentStringFlags32Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentStringFlags32Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		if n, ok := v.value.(uint32); ok {
			return MakeContentStringFlags32Map(m.tree.Insert(k, n)), nil
		}

		return nil, newInvalidContentStringFlags32MapValueError(v)
	}

	return nil, newInvalidContentValueError(value)
}

func (m ContentStringFlags32Map) del(key AttributeValue) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(k)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentStringFlags32Map(t), nil
}

// ContentStringFlags64Map implements ContentSubItem as map of string
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 64 bits flags.
type ContentStringFlags64Map struct {
	tree *strtree64.Tree
}

// MakeContentStringFlags64Map creates instance of ContentStringFlags64Map
// based on strtree64 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 64 bits flags.
func MakeContentStringFlags64Map(tree *strtree64.Tree) ContentStringFlags64Map {
	return ContentStringFlags64Map{tree: tree}
}

func (m ContentStringFlags64Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	s, err := key.str()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeFlagsValue64(v, t), nil
}

func (m ContentStringFlags64Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentStringFlags64Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		if n, ok := v.value.(uint64); ok {
			return MakeContentStringFlags64Map(m.tree.Insert(k, n)), nil
		}

		return nil, newInvalidContentStringFlags64MapValueError(v)
	}

	return nil, newInvalidContentValueError(value)
}

func (m ContentStringFlags64Map) del(key AttributeValue) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(k)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentStringFlags64Map(t), nil
}

// ContentNetworkMap implements ContentSubItem as map of IP address or network
// to ContentSubItem.
type ContentNetworkMap struct {
	tree *iptree.Tree
}

// MakeContentNetworkMap creates instance of ContentNetworkMap based on iptree
// from github.com/infobloxopen/go-trees. Nodes should be of the same
// ContentSubItem compatible type.
func MakeContentNetworkMap(tree *iptree.Tree) ContentNetworkMap {
	return ContentNetworkMap{tree: tree}
}

func (m ContentNetworkMap) getByAttribute(key AttributeValue) (interface{}, error) {
	if a, err := key.address(); err == nil {
		if v, ok := m.tree.GetByIP(a); ok {
			return v, nil
		}

		return UndefinedValue, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if v, ok := m.tree.GetByNet(n); ok {
			return v, nil
		}

		return UndefinedValue, newMissingValueError()
	}

	return UndefinedValue, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkMap) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return UndefinedValue, err
	}

	return MakeContentValue(v).getValue(UndefinedValue, t)
}

func (m ContentNetworkMap) next(key AttributeValue) (ContentSubItem, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return nil, err
	}

	item, ok := v.(ContentSubItem)
	if !ok {
		return nil, newMapContentSubitemError()
	}

	return item, nil
}

func (m ContentNetworkMap) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if v, ok := value.(ContentValue); ok {
			return MakeContentNetworkMap(m.tree.InsertIP(a, v.value)), nil
		}

		return MakeContentNetworkMap(m.tree.InsertIP(a, value)), nil
	}

	if n, err := key.network(); err == nil {
		if v, ok := value.(ContentValue); ok {
			return MakeContentNetworkMap(m.tree.InsertNet(n, v.value)), nil
		}

		return MakeContentNetworkMap(m.tree.InsertNet(n, value)), nil
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkMap) del(key AttributeValue) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if t, ok := m.tree.DeleteByIP(a); ok {
			return MakeContentNetworkMap(t), nil
		}

		return m, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if t, ok := m.tree.DeleteByNet(n); ok {
			return MakeContentNetworkMap(t), nil
		}

		return m, newMissingValueError()
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

// ContentNetworkFlags8Map implements ContentSubItem as map of string
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 8 bits flags.
type ContentNetworkFlags8Map struct {
	tree *iptree8.Tree
}

// MakeContentNetworkFlags8Map creates instance of ContentNetworkFlags8Map
// based on strtree8 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 8 bits flags.
func MakeContentNetworkFlags8Map(tree *iptree8.Tree) ContentNetworkFlags8Map {
	return ContentNetworkFlags8Map{tree: tree}
}

func (m ContentNetworkFlags8Map) getByAttribute(key AttributeValue) (uint8, error) {
	if a, err := key.address(); err == nil {
		if v, ok := m.tree.GetByIP(a); ok {
			return v, nil
		}

		return 0, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if v, ok := m.tree.GetByNet(n); ok {
			return v, nil
		}

		return 0, newMissingValueError()
	}

	return 0, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkFlags8Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return UndefinedValue, err
	}

	return MakeFlagsValue8(v, t), nil
}

func (m ContentNetworkFlags8Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentNetworkFlags8Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if v, ok := value.(ContentValue); ok {
			if n, ok := v.value.(uint8); ok {
				return MakeContentNetworkFlags8Map(m.tree.InsertIP(a, n)), nil
			}

			return nil, newInvalidContentNetworkFlags8MapValueError(v)
		}

		return nil, newInvalidContentValueError(value)
	}

	if n, err := key.network(); err == nil {
		if v, ok := value.(ContentValue); ok {
			if d, ok := v.value.(uint8); ok {
				return MakeContentNetworkFlags8Map(m.tree.InsertNet(n, d)), nil
			}

			return nil, newInvalidContentNetworkFlags8MapValueError(v)
		}

		return nil, newInvalidContentValueError(value)
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkFlags8Map) del(key AttributeValue) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if t, ok := m.tree.DeleteByIP(a); ok {
			return MakeContentNetworkFlags8Map(t), nil
		}

		return m, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if t, ok := m.tree.DeleteByNet(n); ok {
			return MakeContentNetworkFlags8Map(t), nil
		}

		return m, newMissingValueError()
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

// ContentNetworkFlags16Map implements ContentSubItem as map of string
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 16 bits flags.
type ContentNetworkFlags16Map struct {
	tree *iptree16.Tree
}

// MakeContentNetworkFlags16Map creates instance of ContentNetworkFlags16Map
// based on strtree16 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 16 bits flags.
func MakeContentNetworkFlags16Map(tree *iptree16.Tree) ContentNetworkFlags16Map {
	return ContentNetworkFlags16Map{tree: tree}
}

func (m ContentNetworkFlags16Map) getByAttribute(key AttributeValue) (uint16, error) {
	if a, err := key.address(); err == nil {
		if v, ok := m.tree.GetByIP(a); ok {
			return v, nil
		}

		return 0, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if v, ok := m.tree.GetByNet(n); ok {
			return v, nil
		}

		return 0, newMissingValueError()
	}

	return 0, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkFlags16Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return UndefinedValue, err
	}

	return MakeFlagsValue16(v, t), nil
}

func (m ContentNetworkFlags16Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentNetworkFlags16Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if v, ok := value.(ContentValue); ok {
			if n, ok := v.value.(uint16); ok {
				return MakeContentNetworkFlags16Map(m.tree.InsertIP(a, n)), nil
			}

			return nil, newInvalidContentNetworkFlags16MapValueError(v)
		}

		return nil, newInvalidContentValueError(value)
	}

	if n, err := key.network(); err == nil {
		if v, ok := value.(ContentValue); ok {
			if d, ok := v.value.(uint16); ok {
				return MakeContentNetworkFlags16Map(m.tree.InsertNet(n, d)), nil
			}

			return nil, newInvalidContentNetworkFlags16MapValueError(v)
		}

		return nil, newInvalidContentValueError(value)
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkFlags16Map) del(key AttributeValue) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if t, ok := m.tree.DeleteByIP(a); ok {
			return MakeContentNetworkFlags16Map(t), nil
		}

		return m, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if t, ok := m.tree.DeleteByNet(n); ok {
			return MakeContentNetworkFlags16Map(t), nil
		}

		return m, newMissingValueError()
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

// ContentNetworkFlags32Map implements ContentSubItem as map of string
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 32 bits flags.
type ContentNetworkFlags32Map struct {
	tree *iptree32.Tree
}

// MakeContentNetworkFlags32Map creates instance of ContentNetworkFlags32Map
// based on strtree32 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 32 bits flags.
func MakeContentNetworkFlags32Map(tree *iptree32.Tree) ContentNetworkFlags32Map {
	return ContentNetworkFlags32Map{tree: tree}
}

func (m ContentNetworkFlags32Map) getByAttribute(key AttributeValue) (uint32, error) {
	if a, err := key.address(); err == nil {
		if v, ok := m.tree.GetByIP(a); ok {
			return v, nil
		}

		return 0, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if v, ok := m.tree.GetByNet(n); ok {
			return v, nil
		}

		return 0, newMissingValueError()
	}

	return 0, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkFlags32Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return UndefinedValue, err
	}

	return MakeFlagsValue32(v, t), nil
}

func (m ContentNetworkFlags32Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentNetworkFlags32Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if v, ok := value.(ContentValue); ok {
			if n, ok := v.value.(uint32); ok {
				return MakeContentNetworkFlags32Map(m.tree.InsertIP(a, n)), nil
			}

			return nil, newInvalidContentNetworkFlags32MapValueError(v)
		}

		return nil, newInvalidContentValueError(value)
	}

	if n, err := key.network(); err == nil {
		if v, ok := value.(ContentValue); ok {
			if d, ok := v.value.(uint32); ok {
				return MakeContentNetworkFlags32Map(m.tree.InsertNet(n, d)), nil
			}

			return nil, newInvalidContentNetworkFlags32MapValueError(v)
		}

		return nil, newInvalidContentValueError(value)
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkFlags32Map) del(key AttributeValue) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if t, ok := m.tree.DeleteByIP(a); ok {
			return MakeContentNetworkFlags32Map(t), nil
		}

		return m, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if t, ok := m.tree.DeleteByNet(n); ok {
			return MakeContentNetworkFlags32Map(t), nil
		}

		return m, newMissingValueError()
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

// ContentNetworkFlags64Map implements ContentSubItem as map of string
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 64 bits flags.
type ContentNetworkFlags64Map struct {
	tree *iptree64.Tree
}

// MakeContentNetworkFlags64Map creates instance of ContentNetworkFlags64Map
// based on strtree64 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 64 bits flags.
func MakeContentNetworkFlags64Map(tree *iptree64.Tree) ContentNetworkFlags64Map {
	return ContentNetworkFlags64Map{tree: tree}
}

func (m ContentNetworkFlags64Map) getByAttribute(key AttributeValue) (uint64, error) {
	if a, err := key.address(); err == nil {
		if v, ok := m.tree.GetByIP(a); ok {
			return v, nil
		}

		return 0, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if v, ok := m.tree.GetByNet(n); ok {
			return v, nil
		}

		return 0, newMissingValueError()
	}

	return 0, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkFlags64Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return UndefinedValue, err
	}

	return MakeFlagsValue64(v, t), nil
}

func (m ContentNetworkFlags64Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentNetworkFlags64Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if v, ok := value.(ContentValue); ok {
			if n, ok := v.value.(uint64); ok {
				return MakeContentNetworkFlags64Map(m.tree.InsertIP(a, n)), nil
			}

			return nil, newInvalidContentNetworkFlags64MapValueError(v)
		}

		return nil, newInvalidContentValueError(value)
	}

	if n, err := key.network(); err == nil {
		if v, ok := value.(ContentValue); ok {
			if d, ok := v.value.(uint64); ok {
				return MakeContentNetworkFlags64Map(m.tree.InsertNet(n, d)), nil
			}

			return nil, newInvalidContentNetworkFlags64MapValueError(v)
		}

		return nil, newInvalidContentValueError(value)
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkFlags64Map) del(key AttributeValue) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if t, ok := m.tree.DeleteByIP(a); ok {
			return MakeContentNetworkFlags64Map(t), nil
		}

		return m, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if t, ok := m.tree.DeleteByNet(n); ok {
			return MakeContentNetworkFlags64Map(t), nil
		}

		return m, newMissingValueError()
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

// ContentDomainMap implements ContentSubItem as map of domain name
// to ContentSubItem.
type ContentDomainMap struct {
	tree *domaintree.Node
}

// MakeContentDomainMap creates instance of ContentDomainMap based on domaintree
// from github.com/infobloxopen/go-trees. Nodes should be of the same
// ContentSubItem compatible type.
func MakeContentDomainMap(tree *domaintree.Node) ContentDomainMap {
	return ContentDomainMap{tree: tree}
}

func (m ContentDomainMap) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	d, err := key.domain()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeContentValue(v).getValue(UndefinedValue, t)
}

func (m ContentDomainMap) next(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return nil, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return nil, newMissingValueError()
	}

	item, ok := v.(ContentSubItem)
	if !ok {
		return nil, newMapContentSubitemError()
	}

	return item, nil
}

func (m ContentDomainMap) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		return MakeContentDomainMap(m.tree.Insert(d, v.value)), nil
	}

	return MakeContentDomainMap(m.tree.Insert(d, value)), nil
}

func (m ContentDomainMap) del(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(d)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentDomainMap(t), nil
}

// ContentDomainFlags8Map implements ContentSubItem as map of domain name
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 8 bits flags.
type ContentDomainFlags8Map struct {
	tree *domaintree8.Node
}

// MakeContentDomainFlags8Map creates instance of ContentDomainFlags8Map
// based on domaintree8 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 8 bits flags.
func MakeContentDomainFlags8Map(tree *domaintree8.Node) ContentDomainFlags8Map {
	return ContentDomainFlags8Map{tree: tree}
}

func (m ContentDomainFlags8Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	d, err := key.domain()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeFlagsValue8(v, t), nil
}

func (m ContentDomainFlags8Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentDomainFlags8Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		if n, ok := v.value.(uint8); ok {
			return MakeContentDomainFlags8Map(m.tree.Insert(d, n)), nil
		}

		return nil, newInvalidContentDomainFlags8MapValueError(v)
	}

	return nil, newInvalidContentValueError(value)
}

func (m ContentDomainFlags8Map) del(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(d)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentDomainFlags8Map(t), nil
}

// ContentDomainFlags16Map implements ContentSubItem as map of domain name
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 16 bits flags.
type ContentDomainFlags16Map struct {
	tree *domaintree16.Node
}

// MakeContentDomainFlags16Map creates instance of ContentDomainFlags16Map
// based on domaintree16 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 16 bits flags.
func MakeContentDomainFlags16Map(tree *domaintree16.Node) ContentDomainFlags16Map {
	return ContentDomainFlags16Map{tree: tree}
}

func (m ContentDomainFlags16Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	d, err := key.domain()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeFlagsValue16(v, t), nil
}

func (m ContentDomainFlags16Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentDomainFlags16Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		if n, ok := v.value.(uint16); ok {
			return MakeContentDomainFlags16Map(m.tree.Insert(d, n)), nil
		}

		return nil, newInvalidContentDomainFlags16MapValueError(v)
	}

	return nil, newInvalidContentValueError(value)
}

func (m ContentDomainFlags16Map) del(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(d)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentDomainFlags16Map(t), nil
}

// ContentDomainFlags32Map implements ContentSubItem as map of domain name
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 32 bits flags.
type ContentDomainFlags32Map struct {
	tree *domaintree32.Node
}

// MakeContentDomainFlags32Map creates instance of ContentDomainFlags32Map
// based on domaintree32 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 32 bits flags.
func MakeContentDomainFlags32Map(tree *domaintree32.Node) ContentDomainFlags32Map {
	return ContentDomainFlags32Map{tree: tree}
}

func (m ContentDomainFlags32Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	d, err := key.domain()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeFlagsValue32(v, t), nil
}

func (m ContentDomainFlags32Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentDomainFlags32Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		if n, ok := v.value.(uint32); ok {
			return MakeContentDomainFlags32Map(m.tree.Insert(d, n)), nil
		}

		return nil, newInvalidContentDomainFlags32MapValueError(v)
	}

	return nil, newInvalidContentValueError(value)
}

func (m ContentDomainFlags32Map) del(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(d)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentDomainFlags32Map(t), nil
}

// ContentDomainFlags64Map implements ContentSubItem as map of domain name
// to ContentSubItem. In the case resulting ContentSubItem can be only
// a ContentValue instance which holds 64 bits flags.
type ContentDomainFlags64Map struct {
	tree *domaintree64.Node
}

// MakeContentDomainFlags64Map creates instance of ContentDomainFlags64Map
// based on domaintree64 from github.com/infobloxopen/go-trees. Nodes should be
// of the same ContentSubItem compatible type wrapping 64 bits flags.
func MakeContentDomainFlags64Map(tree *domaintree64.Node) ContentDomainFlags64Map {
	return ContentDomainFlags64Map{tree: tree}
}

func (m ContentDomainFlags64Map) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	d, err := key.domain()
	if err != nil {
		return UndefinedValue, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return UndefinedValue, newMissingValueError()
	}

	return MakeFlagsValue64(v, t), nil
}

func (m ContentDomainFlags64Map) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (m ContentDomainFlags64Map) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	if v, ok := value.(ContentValue); ok {
		if n, ok := v.value.(uint64); ok {
			return MakeContentDomainFlags64Map(m.tree.Insert(d, n)), nil
		}

		return nil, newInvalidContentDomainFlags64MapValueError(v)
	}

	return nil, newInvalidContentValueError(value)
}

func (m ContentDomainFlags64Map) del(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(d)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentDomainFlags64Map(t), nil
}

// ContentValue implements ContentSubItem as immediate value.
type ContentValue struct {
	value interface{}
}

// MakeContentValue creates instance of ContentValue with given data. Argument
// value should be value of golang type which corresponds to one of supported
// attribute types.
func MakeContentValue(value interface{}) ContentValue {
	return ContentValue{value: value}
}

func (v ContentValue) getValue(key AttributeValue, t Type) (AttributeValue, error) {
	switch t {
	case TypeUndefined:
		panic(fmt.Errorf("can't convert to value of undefined type"))

	case TypeBoolean:
		return MakeBooleanValue(v.value.(bool)), nil

	case TypeString:
		return MakeStringValue(v.value.(string)), nil

	case TypeInteger:
		return MakeIntegerValue(v.value.(int64)), nil

	case TypeFloat:
		return MakeFloatValue(v.value.(float64)), nil

	case TypeAddress:
		return MakeAddressValue(v.value.(net.IP)), nil

	case TypeNetwork:
		return MakeNetworkValue(v.value.(*net.IPNet)), nil

	case TypeDomain:
		return MakeDomainValue(v.value.(domain.Name)), nil

	case TypeSetOfStrings:
		return MakeSetOfStringsValue(v.value.(*strtree.Tree)), nil

	case TypeSetOfNetworks:
		return MakeSetOfNetworksValue(v.value.(*iptree.Tree)), nil

	case TypeSetOfDomains:
		return MakeSetOfDomainsValue(v.value.(*domaintree.Node)), nil

	case TypeListOfStrings:
		return MakeListOfStringsValue(v.value.([]string)), nil
	}

	panic(fmt.Errorf("can't convert to value of unknown type with index %d", t))
}

func (v ContentValue) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func (v ContentValue) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	return v, newInvalidContentValueModificationError()
}

func (v ContentValue) del(key AttributeValue) (ContentSubItem, error) {
	return v, newInvalidContentValueModificationError()
}
