package pdp

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

// ContentKeyTypes gathers all types which can be a key for content map.
var ContentKeyTypes = map[int]bool{
	TypeString:  true,
	TypeAddress: true,
	TypeNetwork: true,
	TypeDomain:  true,
}

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

	return &LocalContentStorageTransaction{tag: *tag, ID: cID, items: c.items}, nil
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
	tag   uuid.UUID
	ID    string
	items *strtree.Tree
	err   error
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
	id    string
	tag   *uuid.UUID
	items *strtree.Tree
}

// NewLocalContent creates content of given id with given tag and set of content
// items. Nil tag makes the content untagged. Such content can't be
// incrementally updated.
func NewLocalContent(id string, tag *uuid.UUID, items []*ContentItem) *LocalContent {
	c := &LocalContent{id: id, tag: tag, items: strtree.NewTree()}

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
	t  int
	k  []int
}

// MakeContentValueItem creates content item which represents immediate value
// of given type.
func MakeContentValueItem(id string, t int, v interface{}) *ContentItem {
	return &ContentItem{
		id: id,
		r:  MakeContentValue(v),
		t:  t}
}

// MakeContentMappingItem creates mapping content item. Argument t is type
// of final value while k list is a list of types from ContentKeyTypes and
// defines which maps the item consists from.
func MakeContentMappingItem(id string, t int, k []int, v ContentSubItem) *ContentItem {
	return &ContentItem{
		id: id,
		r:  v,
		t:  t,
		k:  k}
}

// GetType returns content item type
func (c *ContentItem) GetType() int {
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

	switch c.t {
	default:
		return nil, newUnknownContentItemResultTypeError(c.t)

	case TypeUndefined:
		return nil, newInvalidContentItemResultTypeError(c.t)

	case TypeBoolean:
		if _, ok := subItem.value.(bool); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeBoolean)
		}

	case TypeInteger:
		if _, ok := subItem.value.(int64); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeInteger)
		}

	case TypeFloat:
		if _, ok := subItem.value.(float64); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeFloat)
		}

	case TypeString:
		if _, ok := subItem.value.(string); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeString)
		}

	case TypeAddress:
		if _, ok := subItem.value.(net.IP); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeAddress)
		}

	case TypeNetwork:
		if _, ok := subItem.value.(*net.IPNet); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeNetwork)
		}

	case TypeDomain:
		if _, ok := subItem.value.(string); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeDomain)
		}

	case TypeSetOfStrings:
		if _, ok := subItem.value.(*strtree.Tree); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeSetOfStrings)
		}

	case TypeSetOfNetworks:
		if _, ok := subItem.value.(*iptree.Tree); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeSetOfNetworks)
		}

	case TypeSetOfDomains:
		if _, ok := subItem.value.(*domaintree.Node); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeSetOfDomains)
		}

	case TypeListOfStrings:
		if _, ok := subItem.value.([]string); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeListOfStrings)
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
		return undefinedValue, newInvalidSelectorPathError(c.k, path)
	}

	if d > 0 {
		m := c.r
		loc := []string{""}
		for _, e := range path[:d-1] {
			key, err := e.Calculate(ctx)
			if err != nil {
				return undefinedValue, bindError(err, strings.Join(loc, "/"))
			}

			loc = append(loc, key.describe())

			m, err = m.next(key)
			if err != nil {
				return undefinedValue, bindError(err, strings.Join(loc, "/"))
			}
		}

		key, err := path[d-1].Calculate(ctx)
		if err != nil {
			return undefinedValue, bindError(err, strings.Join(loc, "/"))
		}

		v, err := m.getValue(key, c.t)
		if err != nil {
			return undefinedValue, bindError(err, strings.Join(append(loc, key.describe()), "/"))
		}

		return v, nil
	}

	return c.r.getValue(undefinedValue, c.t)
}

// ContentSubItem interface abstracts all possible mapping objects and immediate
// content value.
type ContentSubItem interface {
	getValue(key AttributeValue, t int) (AttributeValue, error)
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

func (m ContentStringMap) getValue(key AttributeValue, t int) (AttributeValue, error) {
	s, err := key.str()
	if err != nil {
		return undefinedValue, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return undefinedValue, newMissingValueError()
	}

	return MakeContentValue(v).getValue(undefinedValue, t)
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

		return undefinedValue, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if v, ok := m.tree.GetByNet(n); ok {
			return v, nil
		}

		return undefinedValue, newMissingValueError()
	}

	return undefinedValue, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m ContentNetworkMap) getValue(key AttributeValue, t int) (AttributeValue, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return undefinedValue, err
	}

	return MakeContentValue(v).getValue(undefinedValue, t)
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

func (m ContentDomainMap) getValue(key AttributeValue, t int) (AttributeValue, error) {
	d, err := key.domain()
	if err != nil {
		return undefinedValue, err
	}

	v, ok, err := m.tree.WireGet(d)
	if err != nil {
		return undefinedValue, err
	}

	if !ok {
		return undefinedValue, newMissingValueError()
	}

	return MakeContentValue(v).getValue(undefinedValue, t)
}

func (m ContentDomainMap) next(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return nil, err
	}

	v, ok, err := m.tree.WireGet(d)
	if err != nil {
		return nil, err
	}

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
		return MakeContentDomainMap(m.tree.Insert(d.String(), v.value)), nil
	}

	return MakeContentDomainMap(m.tree.Insert(d.String(), value)), nil
}

func (m ContentDomainMap) del(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(d.String())
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentDomainMap(t), nil
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

func (v ContentValue) getValue(key AttributeValue, t int) (AttributeValue, error) {
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
		return MakeDomainValue(v.value.(domaintree.WireDomainNameLower)), nil

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
