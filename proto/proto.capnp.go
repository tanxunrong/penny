package proto

// AUTO GENERATED - DO NOT EDIT

import (
	C "github.com/glycerine/go-capnproto"
	"unsafe"
)

type Item C.Struct
type Item_Which uint16

const (
	ITEM_NUM Item_Which = 0
	ITEM_STR Item_Which = 1
)

func NewItem(s *C.Segment) Item      { return Item(s.NewStruct(8, 1)) }
func NewRootItem(s *C.Segment) Item  { return Item(s.NewRootStruct(8, 1)) }
func AutoNewItem(s *C.Segment) Item  { return Item(s.NewStructAR(8, 1)) }
func ReadRootItem(s *C.Segment) Item { return Item(s.Root(0).ToStruct()) }
func (s Item) Which() Item_Which     { return Item_Which(C.Struct(s).Get16(4)) }
func (s Item) Num() int32            { return int32(C.Struct(s).Get32(0)) }
func (s Item) SetNum(v int32)        { C.Struct(s).Set16(4, 0); C.Struct(s).Set32(0, uint32(v)) }
func (s Item) Str() string           { return C.Struct(s).GetObject(0).ToText() }
func (s Item) SetStr(v string) {
	C.Struct(s).Set16(4, 1)
	C.Struct(s).SetObject(0, s.Segment.NewText(v))
}

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s Item) MarshalJSON() (bs []byte, err error) { return }

type Item_List C.PointerList

func NewItemList(s *C.Segment, sz int) Item_List { return Item_List(s.NewCompositeList(8, 1, sz)) }
func (s Item_List) Len() int                     { return C.PointerList(s).Len() }
func (s Item_List) At(i int) Item                { return Item(C.PointerList(s).At(i).ToStruct()) }
func (s Item_List) ToArray() []Item              { return *(*[]Item)(unsafe.Pointer(C.PointerList(s).ToArray())) }
func (s Item_List) Set(i int, item Item)         { C.PointerList(s).Set(i, C.Object(item)) }

type Pair C.Struct
type PairVal Pair
type PairVal_Which uint16

const (
	PAIRVAL_VALUE PairVal_Which = 0
	PAIRVAL_SUB   PairVal_Which = 1
)

func NewPair(s *C.Segment) Pair        { return Pair(s.NewStruct(8, 2)) }
func NewRootPair(s *C.Segment) Pair    { return Pair(s.NewRootStruct(8, 2)) }
func AutoNewPair(s *C.Segment) Pair    { return Pair(s.NewStructAR(8, 2)) }
func ReadRootPair(s *C.Segment) Pair   { return Pair(s.Root(0).ToStruct()) }
func (s Pair) Key() Item               { return Item(C.Struct(s).GetObject(0).ToStruct()) }
func (s Pair) SetKey(v Item)           { C.Struct(s).SetObject(0, C.Object(v)) }
func (s Pair) Val() PairVal            { return PairVal(s) }
func (s PairVal) Which() PairVal_Which { return PairVal_Which(C.Struct(s).Get16(0)) }
func (s PairVal) Value() Item          { return Item(C.Struct(s).GetObject(1).ToStruct()) }
func (s PairVal) SetValue(v Item)      { C.Struct(s).Set16(0, 0); C.Struct(s).SetObject(1, C.Object(v)) }
func (s PairVal) Sub() Pair_List       { return Pair_List(C.Struct(s).GetObject(1)) }
func (s PairVal) SetSub(v Pair_List)   { C.Struct(s).Set16(0, 1); C.Struct(s).SetObject(1, C.Object(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s Pair) MarshalJSON() (bs []byte, err error) { return }

type Pair_List C.PointerList

func NewPairList(s *C.Segment, sz int) Pair_List { return Pair_List(s.NewCompositeList(8, 2, sz)) }
func (s Pair_List) Len() int                     { return C.PointerList(s).Len() }
func (s Pair_List) At(i int) Pair                { return Pair(C.PointerList(s).At(i).ToStruct()) }
func (s Pair_List) ToArray() []Pair              { return *(*[]Pair)(unsafe.Pointer(C.PointerList(s).ToArray())) }
func (s Pair_List) Set(i int, item Pair)         { C.PointerList(s).Set(i, C.Object(item)) }

type Param C.Struct
type Param_Which uint16

const (
	PARAM_ITEM Param_Which = 0
	PARAM_PAIR Param_Which = 1
)

func NewParam(s *C.Segment) Param      { return Param(s.NewStruct(8, 1)) }
func NewRootParam(s *C.Segment) Param  { return Param(s.NewRootStruct(8, 1)) }
func AutoNewParam(s *C.Segment) Param  { return Param(s.NewStructAR(8, 1)) }
func ReadRootParam(s *C.Segment) Param { return Param(s.Root(0).ToStruct()) }
func (s Param) Which() Param_Which     { return Param_Which(C.Struct(s).Get16(0)) }
func (s Param) Item() Item             { return Item(C.Struct(s).GetObject(0).ToStruct()) }
func (s Param) SetItem(v Item)         { C.Struct(s).Set16(0, 0); C.Struct(s).SetObject(0, C.Object(v)) }
func (s Param) Pair() Pair_List        { return Pair_List(C.Struct(s).GetObject(0)) }
func (s Param) SetPair(v Pair_List)    { C.Struct(s).Set16(0, 1); C.Struct(s).SetObject(0, C.Object(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s Param) MarshalJSON() (bs []byte, err error) { return }

type Param_List C.PointerList

func NewParamList(s *C.Segment, sz int) Param_List { return Param_List(s.NewCompositeList(8, 1, sz)) }
func (s Param_List) Len() int                      { return C.PointerList(s).Len() }
func (s Param_List) At(i int) Param                { return Param(C.PointerList(s).At(i).ToStruct()) }
func (s Param_List) ToArray() []Param              { return *(*[]Param)(unsafe.Pointer(C.PointerList(s).ToArray())) }
func (s Param_List) Set(i int, item Param)         { C.PointerList(s).Set(i, C.Object(item)) }

type Msg C.Struct

func NewMsg(s *C.Segment) Msg        { return Msg(s.NewStruct(8, 4)) }
func NewRootMsg(s *C.Segment) Msg    { return Msg(s.NewRootStruct(8, 4)) }
func AutoNewMsg(s *C.Segment) Msg    { return Msg(s.NewStructAR(8, 4)) }
func ReadRootMsg(s *C.Segment) Msg   { return Msg(s.Root(0).ToStruct()) }
func (s Msg) From() string           { return C.Struct(s).GetObject(0).ToText() }
func (s Msg) SetFrom(v string)       { C.Struct(s).SetObject(0, s.Segment.NewText(v)) }
func (s Msg) Dest() string           { return C.Struct(s).GetObject(1).ToText() }
func (s Msg) SetDest(v string)       { C.Struct(s).SetObject(1, s.Segment.NewText(v)) }
func (s Msg) Pass() int32            { return int32(C.Struct(s).Get32(0)) }
func (s Msg) SetPass(v int32)        { C.Struct(s).Set32(0, uint32(v)) }
func (s Msg) Method() string         { return C.Struct(s).GetObject(2).ToText() }
func (s Msg) SetMethod(v string)     { C.Struct(s).SetObject(2, s.Segment.NewText(v)) }
func (s Msg) Params() Param_List     { return Param_List(C.Struct(s).GetObject(3)) }
func (s Msg) SetParams(v Param_List) { C.Struct(s).SetObject(3, C.Object(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s Msg) MarshalJSON() (bs []byte, err error) { return }

type Msg_List C.PointerList

func NewMsgList(s *C.Segment, sz int) Msg_List { return Msg_List(s.NewCompositeList(8, 4, sz)) }
func (s Msg_List) Len() int                    { return C.PointerList(s).Len() }
func (s Msg_List) At(i int) Msg                { return Msg(C.PointerList(s).At(i).ToStruct()) }
func (s Msg_List) ToArray() []Msg              { return *(*[]Msg)(unsafe.Pointer(C.PointerList(s).ToArray())) }
func (s Msg_List) Set(i int, item Msg)         { C.PointerList(s).Set(i, C.Object(item)) }
