package rdb

// token represents an atomic component of an SQL statements
// For RDB purposes we're only interested in column and table identifiers
// and aliases so that datatype and data origination/destination can be established.
type token int

// The following are a subset of MySQL tokens
const (
	illegal token = iota // Illegal or unrecognized token

	EOF                 // End of file
	WS                  // Whitespace
	unparsed            // Unparsed text
	naturalNumber       // 0, 1, 2, ..., n
	integer             // ..., -1, 0, 1, ...
	fixedNumber         // 3.1415
	floatingPointNumber // 3.233e9
	identifier          // named column, table, alias, procedure, variable, etc
	quotedString        // a string literal
	astrisk             // *
	comma               // ,
	period              // .
	lParen              // (
	rParen              // )
	lBrace              // {
	rBrace              // }
	equals              // =
	selectToken
	insertToken
	fromToken
	partitionToken
	asToken
	straightJoinToken
	crossJoinToken
	innerJoinToken
	ojToken
	naturalJoinToken
	naturalLeftJoinToken
	naturalLeftOuterJoinToken
	naturalRightJoinToken
	naturalRightOuterJoinToken
	leftJoinToken
	leftOuterJoinToken
	rightJoinToken
	rightOuterJoinToken
	useIndexToken
	useKeyToken
	ignoreIndexToken
	ignoreKeyToken
	forceIndexToken
	forceKeyToken
	forJoinToken
	forOrderByToken
	forGroupByToken
	whereToken
	valuesToken
	setToken
	defaultToken
	allToken
	distinctToken
	highPriorityToken
	lowPriorityToken
	delayedToken
	maxStatementTimeToken
	sqlSmallResultToken
	sqlBigResultToken
	sqlBufferResultToken
	sqlCacheToken
	sqlNoCacheToken
	sqlCalcFoundRowsToken
	onToken
	usingToken
	orderByToken
	groupByToken
)

const (
	selectStmt                = "SELECT"
	insertStmt                = "INSERT"
	fromStmt                  = "FROM"
	partitionStmt             = "PARTITION"
	asStmt                    = "AS"
	straightJoinStmt          = "STRAIGHT_JOIN"
	crossJoinStmt             = "CROSS JOIN"
	innerJoinStmt             = "INNER JOIN"
	ojStmt                    = "OJ"
	naturalJoinStmt           = "NATURAL JOIN"
	naturalLeftJoinStmt       = "NATURAL LEFT JOIN"
	naturalLeftOuterJoinStmt  = "NATURAL LEFT OUTER JOIN"
	naturalRightJoinStmt      = "NATURAL RIGHT JOIN"
	naturalRightOuterJoinStmt = "NATURAL RIGHT OUTER JOIN"
	leftJoinStmt              = "LEFT JOIN"
	leftOuterJoinStmt         = "LEFT OUTER JOIN"
	rightJoinStmt             = "RIGHT JOIN"
	rightOuterJoinStmt        = "RIGHT OUTER JOIN"
	useIndexStmt              = "USE INDEX"
	useKeyStmt                = "USE KEY"
	ignoreIndexStmt           = "IGNORE INDEX"
	ignoreKeyStmt             = "IGNORE KEY"
	forceIndexStmt            = "FORCE INDEX"
	forceKeyStmt              = "FORCE KEY"
	forJoinStmt               = "FOR JOIN"
	forOrderByStmt            = "FOR ORDER BY"
	forGroupByStmt            = "FOR GROUP BY"
	whereStmt                 = "WHERE"
	valuesStmt                = "VALUES"
	setStmt                   = "SET"
	defaultStmt               = "DEFAULT"
	allStmt                   = "ALL"
	distinctStmt              = "DISTINCT"
	highPriorityStmt          = "HIGH_PRIORITY"
	lowPriorityStmt           = "LOW_PRIORITY"
	delayedStmt               = "DELAYED"
	maxStatementTimeStmt      = "MAX_STATEMENT_TIME"
	sqlSmallResultStmt        = "SQL_SMALL_RESULT"
	sqlBigResultStmt          = "SQL_BIG_RESULT"
	sqlBufferResultStmt       = "SQL_BUFFER_RESULT"
	sqlCacheStmt              = "SQL_CACHE"
	sqlNoCacheStmt            = "SQL_NO_CACHE"
	sqlCalcFoundRowsStmt      = "SQL_CALC_FOUND_ROWS"
	onStmt                    = "ON"
	usingStmt                 = "USING"
	orderByStmt               = "ORDER BY"
	groupByStmt               = "GROUP BY"
)
