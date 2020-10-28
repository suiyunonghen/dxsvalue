package dxsvalue

type BsonType	uint8

const(
	BSON_Double           BsonType = 0x01     //64-bit binary floating point
	BSON_String           BsonType = 0x02	  //UTF-8 string
	BSON_EmbeddedDocument BsonType = 0x03	  //Embedded document
	BSON_Array            BsonType = 0x04	  //Array
	BSON_Binary           BsonType = 0x05	  //Binary data ,具备有子类型
	BSON_Undefined        BsonType = 0x06
	BSON_ObjectID         BsonType = 0x07
	BSON_Boolean          BsonType = 0x08
	BSON_DateTime         BsonType = 0x09
	BSON_Null             BsonType = 0x0A
	BSON_Regex            BsonType = 0x0B
	BSON_DBPointer        BsonType = 0x0C	  //DBPointer — Deprecated
	BSON_JavaScript       BsonType = 0x0D
	BSON_Symbol           BsonType = 0x0E	  //Symbol. — Deprecated
	BSON_CodeWithScope    BsonType = 0x0F	  //JavaScript code w/ scope — Deprecated
	BSON_Int32            BsonType = 0x10
	BSON_Timestamp        BsonType = 0x11
	BSON_Int64            BsonType = 0x12
	BSON_Decimal128       BsonType = 0x13
	BSON_MinKey           BsonType = 0xFF
	BSON_MaxKey           BsonType = 0x7F
)

func (bsonType BsonType)String()string  {
	switch bsonType {
		case '\x01':
			return "double"
		case '\x02':
			return "string"
		case '\x03':
			return "embedded document"
		case '\x04':
			return "array"
		case '\x05':
			return "binary"
		case '\x06':
			return "undefined"
		case '\x07':
			return "objectID"
		case '\x08':
			return "boolean"
		case '\x09':
			return "UTC datetime"
		case '\x0A':
			return "null"
		case '\x0B':
			return "regex"
		case '\x0C':
			return "dbPointer"
		case '\x0D':
			return "javascript"
		case '\x0E':
			return "symbol"
		case '\x0F':
			return "code with scope"
		case '\x10':
			return "32-bit integer"
		case '\x11':
			return "timestamp"
		case '\x12':
			return "64-bit integer"
		case '\x13':
			return "128-bit decimal"
		case '\xFF':
			return "min key"
		case '\x7F':
			return "max key"
		default:
			return "invalid"
	}
}

type BsonBinaryType	uint8
const(
	BinaryGeneric     BsonBinaryType = iota		//普通二进制
	BinaryFunction								//函数
	BinaryBinaryOld
	BinaryUUIDOld
	BinaryUUID
	BinaryMD5
	BinaryUserDefined BsonBinaryType = 0x80
)