package main

import (
	"fmt"
	"math"
	"strings"
)

// main routine
func main() {

	b := make(Benchmark)
	_ = b
	//runRealDataEvaluation(10, 0, true)
	runForVariables(b)
	//runSimulation(b, 10)
	// printPerformance(b)
}

func runSimulation(b Benchmark, x float64) Result {

	p := newParameters(x)
	defer b.track(runningtime("TSA=" + strings.ToUpper(p.TSA) + ", X=" + fmt.Sprintf("%.2f", x) + ", " + "\tTime"))
	c := make(chan bool, p.nParallelSims)
	r := make([]Result, p.nParallelSims)
	var f Result

	for i := 0; i < p.nParallelSims; i++ {
		p.Seed = int64(i*p.nRun + 1)
		go run(p, &r[i], c)
	}
	for i := 0; i < p.nParallelSims; i++ {
		<-c
	}

	for _, batch := range r {
		f.JoinResults(batch, p)
	}

	fmt.Println("\nTSA=", strings.ToUpper(p.TSA), "\tLambda=", p.Lambda, "\tD=", p.D, "\tNumberOfNodes=", p.numberNodes, "\tZipf=", p.zipf, "\tTangleSize =", p.TangleSize)
	f.FinalEvaluationSaveResults(p)
	fmt.Println("- - - Confirmation Rate - - -")
	fmt.Println("X\tmean\tSTD ")
	fmt.Println(x, "\t", f.confirmationTime.totalMean, "\t", math.Sqrt(f.confirmationTime.totalVariance))
	return f
}

func run(p Parameters, r *Result, c chan bool) {
	defer func() { c <- true }()
	b := make(Benchmark)
	*r, b = p.RunTangle()
	printPerformance(b)
}

func runForVariables(b Benchmark) {
	var total string
	//Xs := []float64{2, 3, 4, 5, 6, 7, 8, 9, 10}
	Xs := []float64{0, .4, .8, 1.0, 1.2, 1.6, 1.8, 2.0, 3.0}
	//Xs := []float64{1.0}
	//NXs := 2
	//Xs := make([]float64, NXs+1)
	//for i1 := 0; i1 < NXs+1; i1++ {
	//	Xs[i1] = 1. / float64(NXs) * float64(i1)
	//	// Xs[i1] = 5 * (float64(i1) + 1)
	//	// Xs[i1] = 2 + float64(i1)
	//}
	// for i1 := 0; i1 < NXs; i1++ {
	// 	Xs[i1] = .1 * math.Pow(100, float64(i1)/float64(NXs-1))
	// }
	fmt.Println("Variables=", Xs)
	var banner string
	for _, x := range Xs {
		fmt.Println("X=", x)
		r := runSimulation(b, x)
		if banner == "" {
			banner += fmt.Sprintf("#x\tMean ConfirmationTime \tSD \tTips\n")
		}

		output := fmt.Sprintf("%.4f", x)
		output += fmt.Sprintf("\t%.2f", r.confirmationTime.totalMean)
		output += fmt.Sprintf("\t\t\t%.2f", math.Sqrt(r.confirmationTime.totalVariance))
		//output += fmt.Sprintf("\t%.8f", r.tips.STDOrphanTipsRatio)
		output += fmt.Sprintf("\t%.2f", r.tips.tAVG)
		//output += fmt.Sprintf("\t%.8f", r.tips.tSTD)
		output += fmt.Sprintf("\n")

		total += output
		fmt.Println(banner + output)
	}
	fmt.Println(banner + total)
}
