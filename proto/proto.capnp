using Go = import "go.capnp";

$Go.package("proto");

$Go.import("penny/proto");

@0xde88b0bd250ca57a;

struct Item {
	union {
		num   @0:Int32;
		str   @1:Text;
	}
}

struct Pair {
	key @0: Item;
	val :union {
		value @1: Item;
		sub @2: List(Pair);
	}
}

struct Param {
	union {
		item @0: Item;
		pair @1: List(Pair);
	}
}
		
struct Msg {
	from @0:Text;
	dest @1:Text;
	pass @2:Int32;
	method @3:Text;
	params @4:List(Param);
}
