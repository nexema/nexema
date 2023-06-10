package rules

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/analyzer"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

// ValidMapArguments checks if the value type defined as map contains exactly two type arguments and also they are a valid Nexema value type
type ValidMapArguments struct{}

func (self ValidMapArguments) Analyze(context *analyzer.AnalyzerContext) {
	context.RunOver(func(object *scope.Object, source *parser.TypeStmt) {
		for _, stmt := range source.Fields {

			// this should not happen
			if stmt.ValueType == nil {
				panic("this should not happen, field does not have a defined value type?")
			}

			valueType, _ := stmt.ValueType.Format()
			primitive, ok := definition.ParsePrimitive(valueType)

			// checked by other rule
			if !ok {
				continue
			}

			if primitive != definition.Map {
				continue
			}

			length := len(stmt.ValueType.Args)
			if length != 2 {
				context.ReportError(errInvalidMapArgumentsLen{Given: length}, stmt.ValueType.Pos)
				continue
			}

			keyArgument := stmt.ValueType.Args[0]
			valueArgument := stmt.ValueType.Args[1]
			verifyFieldType(&keyArgument, context) // maybe there is a cleaner way?
			verifyFieldType(&valueArgument, context)
		}
	})
}

func (self ValidMapArguments) Throws() analyzer.RuleThrow {
	return analyzer.Error
}

func (self ValidMapArguments) Key() string {
	return "valid-map-arguments"
}

type errInvalidMapArgumentsLen struct {
	Given int
}

func (e errInvalidMapArgumentsLen) Message() string {
	return fmt.Sprintf("map type expects exactly two type arguments, given %d", e.Given)
}
