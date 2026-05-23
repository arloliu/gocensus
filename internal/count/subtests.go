package count

import "go/ast"

type runStats struct {
	staticRuns      int
	dynamicRunSites int
}

type runContext struct {
	multiplier int
	dynamic    bool
}

func firstParamName(fn *ast.FuncDecl) string {
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return ""
	}
	names := fn.Type.Params.List[0].Names
	if len(names) == 0 {
		return ""
	}
	return names[0].Name
}

func runStatsFor(fn *ast.FuncDecl, receiver string) runStats {
	if fn.Body == nil || receiver == "" {
		return runStats{}
	}
	tables := staticTableCounts(fn.Body)
	return runStatsInBlock(fn.Body.List, receiver, tables, runContext{multiplier: 1})
}

func staticTableCounts(body *ast.BlockStmt) map[string]int {
	tables := map[string]int{}
	ast.Inspect(body, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.AssignStmt:
			for i, lhs := range stmt.Lhs {
				if i >= len(stmt.Rhs) {
					continue
				}
				name, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}
				if count, ok := compositeLen(stmt.Rhs[i]); ok {
					tables[name.Name] = count
				}
			}
		case *ast.ValueSpec:
			for i, name := range stmt.Names {
				if i >= len(stmt.Values) {
					continue
				}
				if count, ok := compositeLen(stmt.Values[i]); ok {
					tables[name.Name] = count
				}
			}
		}
		return true
	})
	return tables
}

func compositeLen(expr ast.Expr) (int, bool) {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return 0, false
	}
	return len(lit.Elts), true
}

func runStatsInBlock(stmts []ast.Stmt, receiver string, tables map[string]int, ctx runContext) runStats {
	var stats runStats
	for _, stmt := range stmts {
		stats.add(runStatsInStmt(stmt, receiver, tables, ctx))
	}
	return stats
}

func runStatsInStmt(stmt ast.Stmt, receiver string, tables map[string]int, ctx runContext) runStats {
	switch stmt := stmt.(type) {
	case *ast.RangeStmt:
		next := ctx
		if count, ok := rangeStaticCount(stmt.X, tables); ok && !ctx.dynamic {
			next.multiplier *= count
		} else {
			next.dynamic = true
		}
		return runStatsInBlock(stmt.Body.List, receiver, tables, next)
	case *ast.ForStmt:
		next := ctx
		next.dynamic = true
		return runStatsInBlock(stmt.Body.List, receiver, tables, next)
	case *ast.BlockStmt:
		return runStatsInBlock(stmt.List, receiver, tables, ctx)
	case *ast.IfStmt:
		stats := runStatsInBlock(stmt.Body.List, receiver, tables, ctx)
		if stmt.Else != nil {
			stats.add(runStatsInStmt(stmt.Else, receiver, tables, ctx))
		}
		return stats
	case *ast.SwitchStmt:
		return runStatsInCaseClauses(stmt.Body.List, receiver, tables, ctx)
	case *ast.TypeSwitchStmt:
		return runStatsInCaseClauses(stmt.Body.List, receiver, tables, ctx)
	case *ast.SelectStmt:
		return runStatsInCaseClauses(stmt.Body.List, receiver, tables, ctx)
	default:
		return runStatsInNode(stmt, receiver, ctx)
	}
}

func runStatsInCaseClauses(stmts []ast.Stmt, receiver string, tables map[string]int, ctx runContext) runStats {
	var stats runStats
	for _, stmt := range stmts {
		switch clause := stmt.(type) {
		case *ast.CaseClause:
			stats.add(runStatsInBlock(clause.Body, receiver, tables, ctx))
		case *ast.CommClause:
			stats.add(runStatsInBlock(clause.Body, receiver, tables, ctx))
		default:
			stats.add(runStatsInStmt(stmt, receiver, tables, ctx))
		}
	}
	return stats
}

func rangeStaticCount(expr ast.Expr, tables map[string]int) (int, bool) {
	if count, ok := compositeLen(expr); ok {
		return count, true
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return 0, false
	}
	count, ok := tables[ident.Name]
	return count, ok
}

func runStatsInNode(node ast.Node, receiver string, ctx runContext) runStats {
	var stats runStats
	ast.Inspect(node, func(current ast.Node) bool {
		call, ok := current.(*ast.CallExpr)
		if !ok || !isRunCall(call, receiver) {
			return true
		}
		if ctx.dynamic {
			stats.dynamicRunSites++
			return true
		}
		stats.staticRuns += ctx.multiplier
		return true
	})
	return stats
}

func isRunCall(call *ast.CallExpr, receiver string) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Run" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == receiver
}

func (s *runStats) add(other runStats) {
	s.staticRuns += other.staticRuns
	s.dynamicRunSites += other.dynamicRunSites
}
