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

var combinationOperators = map[string]struct{}{
	opAnd: {},
	opOr:  {},
}

func isCombinationOperator(str string) bool {
	_, ok := combinationOperators[str]
	return ok
}

var conditionOperators = map[string]struct{}{
	opLT:  {},
	opLTE: {},
	opEq:  {},
	opNE:  {},
	opGTE: {},
	opGT:  {},
}

func isConditionOperator(str string) bool {
	_, ok := conditionOperators[str]
	return ok
}

var supportedOperators = map[string]struct{}{
	opAnd: {},
	opOr:  {},
	opLT:  {},
	opLTE: {},
	opEq:  {},
	opNE:  {},
	opGTE: {},
	opGT:  {},
}

func isSupportedOperator(str string) bool {
	_, ok := supportedOperators[str]
	return ok
}

func isOperator(str string) bool {
	return str[0] == '$'
}
