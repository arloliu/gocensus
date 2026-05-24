package diff

import (
	"slices"

	"github.com/arloliu/gocensus"
)

// Options identifies the two census results being compared.
type Options struct {
	Root  string
	Base  string
	Head  string
	Scope string
}

// Report contains repository and package deltas between two census results.
type Report struct {
	Root     string         `json:"root"`
	Base     string         `json:"base"`
	Head     string         `json:"head"`
	Scope    string         `json:"scope"`
	Summary  Summary        `json:"summary"`
	Packages []PackageDelta `json:"packages"`
}

// Summary contains repository-level deltas.
type Summary struct {
	TotalFiles          IntDelta   `json:"total_files"`
	ProductionFiles     IntDelta   `json:"production_files"`
	TestFiles           IntDelta   `json:"test_files"`
	ProductionEffective IntDelta   `json:"production_effective"`
	TestEffective       IntDelta   `json:"test_effective"`
	TestToProduction    FloatDelta `json:"test_to_production"`
	TestShare           FloatDelta `json:"test_share"`
}

// PackageDelta contains package-level production and test deltas.
type PackageDelta struct {
	Package             string     `json:"package"`
	ProductionEffective IntDelta   `json:"production_effective"`
	TestEffective       IntDelta   `json:"test_effective"`
	TestToProduction    FloatDelta `json:"test_to_production"`
}

// IntDelta stores base, head, and head-minus-base integer values.
type IntDelta struct {
	Base  int `json:"base"`
	Head  int `json:"head"`
	Delta int `json:"delta"`
}

// FloatDelta stores base, head, and head-minus-base floating point values.
type FloatDelta struct {
	Base  float64 `json:"base"`
	Head  float64 `json:"head"`
	Delta float64 `json:"delta"`
}

// Compare computes repository and package deltas between base and head.
func Compare(opts Options, base gocensus.Result, head gocensus.Result) Report {
	return Report{
		Root:  opts.Root,
		Base:  opts.Base,
		Head:  opts.Head,
		Scope: opts.Scope,
		Summary: Summary{
			TotalFiles:          intDelta(base.Files.Total, head.Files.Total),
			ProductionFiles:     intDelta(base.Files.Production, head.Files.Production),
			TestFiles:           intDelta(base.Files.Tests, head.Files.Tests),
			ProductionEffective: intDelta(base.Lines.Production.Effective, head.Lines.Production.Effective),
			TestEffective:       intDelta(base.Lines.Tests.Effective, head.Lines.Tests.Effective),
			TestToProduction:    floatDelta(base.Ratios.TestToProductionEffective, head.Ratios.TestToProductionEffective),
			TestShare:           floatDelta(base.Ratios.TestShareEffective, head.Ratios.TestShareEffective),
		},
		Packages: comparePackages(base.Packages, head.Packages),
	}
}

func intDelta(base int, head int) IntDelta {
	return IntDelta{Base: base, Head: head, Delta: head - base}
}

func floatDelta(base float64, head float64) FloatDelta {
	return FloatDelta{Base: base, Head: head, Delta: head - base}
}

func comparePackages(base []gocensus.PackageMetric, head []gocensus.PackageMetric) []PackageDelta {
	byPackage := map[string]PackageDelta{}
	for _, pkg := range base {
		delta := byPackage[pkg.ImportPath]
		delta.Package = pkg.ImportPath
		delta.ProductionEffective.Base = pkg.Lines.Production.Effective
		delta.TestEffective.Base = pkg.Lines.Tests.Effective
		delta.TestToProduction.Base = pkg.Ratios.TestToProductionEffective
		byPackage[pkg.ImportPath] = delta
	}
	for _, pkg := range head {
		delta := byPackage[pkg.ImportPath]
		delta.Package = pkg.ImportPath
		delta.ProductionEffective.Head = pkg.Lines.Production.Effective
		delta.TestEffective.Head = pkg.Lines.Tests.Effective
		delta.TestToProduction.Head = pkg.Ratios.TestToProductionEffective
		delta.ProductionEffective.Delta = delta.ProductionEffective.Head - delta.ProductionEffective.Base
		delta.TestEffective.Delta = delta.TestEffective.Head - delta.TestEffective.Base
		delta.TestToProduction.Delta = delta.TestToProduction.Head - delta.TestToProduction.Base
		byPackage[pkg.ImportPath] = delta
	}

	packages := make([]PackageDelta, 0, len(byPackage))
	for _, delta := range byPackage {
		if delta.ProductionEffective.Delta == 0 && delta.TestEffective.Delta == 0 && delta.TestToProduction.Delta == 0 {
			continue
		}
		packages = append(packages, delta)
	}
	slices.SortFunc(packages, func(a PackageDelta, b PackageDelta) int {
		aAbs := abs(a.ProductionEffective.Delta)
		bAbs := abs(b.ProductionEffective.Delta)
		if aAbs > bAbs {
			return -1
		}
		if aAbs < bAbs {
			return 1
		}
		if a.Package < b.Package {
			return -1
		}
		if a.Package > b.Package {
			return 1
		}
		return 0
	})
	return packages
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
