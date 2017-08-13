package mango

const (
	opNone = ""

	// Combination operators - http://docs.couchdb.org/en/2.0.0/api/database/find.html#combination-operators
	opAnd = "$and"
	opOr  = "$or"
	// opNot       = "$not")
	// opNor       = "$nor")
	// opAll       = "$all")
	// opElemMatch = "$elemMatch")

	// Condition operators - http://docs.couchdb.org/en/2.0.0/api/database/find.html#condition-operators
	opLT  = "$lt"
	opLTE = "$lte"
	opEq  = "$eq"
	opNE  = "$ne"
	opGTE = "$gte"
	opGT  = "$gt"
	// opExists = "$exists")
	// opType   = "$type")
	// opIn     = "$in")
	// opNIn    = "$nin")
	// opSize   = "$size")
	// opMod    = "$mod")
	// opRegex  = "$regex")
)
