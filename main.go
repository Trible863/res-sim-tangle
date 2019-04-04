package main

import (
	"fmt"
	"log"
	"math"
	"strings"

	"gonum.org/v1/gonum/stat"
)

// var nParallelSims = 1

// factor 2 is to use the physical cores, whereas NumCPU returns double the number due to hyper-threading
// var nParallelSims = runtime.NumCPU()/2 - 1

func main() {

	b := make(Benchmark)
	_ = b
	runRealDataEvaluation(10, 0, true)
	// runSimulation(b, 10, 0)
	// printPerformance(b)
}

func runSimulation(b Benchmark, lambda, alpha float64) Result {

	p := newParameters(lambda, alpha)
	defer b.track(runningtime("TSA=" + strings.ToUpper(p.TSA) + ", Lambda=" + fmt.Sprintf("%.2f", lambda) + ", Alpha=" + fmt.Sprintf("%.4f", alpha) + "\tTime"))
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

	fmt.Println("\nTSA=", strings.ToUpper(p.TSA), "\tLambda=", p.Lambda, "\tAlpha=", p.Alpha)
	//fmt.Println(f.avgtips)
	f.FinalEvaluationSaveResults(p)
	return f
}

func run(p Parameters, r *Result, c chan bool) {
	defer func() { c <- true }()
	b := make(Benchmark)
	*r, b = p.RunTangle()
	printPerformance(b)
}

func runRealDataEvaluation(lambda, alpha float64, pull bool) {
	p := newParameters(lambda, alpha)
	var r Result
	sim := Sim{}
	sim.param = p
	r.initResults(&p)
	sim.clearSim()

	if pull {

		//pull real data from IRI
		var endpoint = "http://35.246.92.25:14265"
		err := pullData("data/trytes.txt", endpoint, 1000)
		if err != nil {
			log.Fatal(err)
			fmt.Println(err)
		}
	}
	//convert trytes to Tangle with []Tx

	var err error
	err = sim.buildTangleFromFile("data/trytes.txt")
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}

	if !isRefConsistent(sim.tangle) {
		fmt.Println("ERROR: Tangle is not ref consistent")
		panic(0)
	}

	fmt.Println("CW comparison:", sim.compareCW())
	fmt.Println("Tangle consistency:", isRefConsistent(sim.tangle))

	//printCWRef(sim.cw)

	saveTangle(sim.tangle)

	fmt.Println("Tangle size", len(sim.tangle))
	fmt.Println("CW size:", len(sim.cw))

	//Visualize the Tangle
	if p.drawTangleMode > 0 {
		sim.visualizeTangle(nil, p.drawTangleMode)
	} else if p.drawTangleMode < 0 {
		sim.visualizeRW()
	}

	//sim.tangle=function(datafile)
	//r.EvaluateTangle(&sim, &p, 0)
	//r.FinalEvaluationSaveResults(p)
}

func runForAlphasLambdas(b Benchmark) string {
	// b := make(Benchmark)
	//var ratio string
	var total string
	// lambdas := []float64{3, 10, 30, 100, 300}
	Nlambdas := 30
	lambdas := make([]float64, Nlambdas)
	for i1 := 0; i1 < Nlambdas; i1++ {
		lambdas[i1] = .1 * math.Pow(3000, float64(i1)/float64(Nlambdas-1))
	}
	alphas := []float64{0}
	// Nalphas := 20
	// alphas := make([]float64, Nalphas)
	// for i1 := 0; i1 < Nalphas; i1++ {
	// 	alphas[i1] = 10. * math.Pow(30000, -float64(i1)/float64(Nalphas))
	// }

	// alphas := []float64{0.}
	var banner string
	for _, lambda := range lambdas {
		for _, alpha := range alphas {
			//for alpha := 0.001; alpha <= 0.1; alpha += 0.001 {
			//for lambda := 1.; lambda <= 100; lambda++ {
			// if (alpha * lambda) < 10 {
			// r := runSimulation(b, "rw", lambda, alpha)
			if lambda > 0 {
				r := runSimulation(b, lambda, alpha)
				if banner == "" {
					banner += fmt.Sprintf("#alpha\t")
					for _, m := range r.velocity.vTime {
						banner += fmt.Sprintf("%v\t", m.desc)
					}
					banner += fmt.Sprintf("OP\tTOP\n")
				}

				output := fmt.Sprintf("%.3f", alpha)
				for _, m := range r.velocity.vTime {
					x, y := r.velocity.getTimeMetric(m.desc)
					output += fmt.Sprintf("\t%.5f", stat.Mean(x, y))
				}
				// output += fmt.Sprintf("\t%.5f", stat.Mean(r.op.op, nil))
				// output += fmt.Sprintf("\t%.5f", stat.Mean(r.op.top, nil))
				output += fmt.Sprintf("\t%.5f", stat.Mean(r.op.op2, nil))
				output += fmt.Sprintf("\t%.5f", stat.Mean(r.op.top2, nil))
				output += fmt.Sprintf("\n")

				total += output
				fmt.Println(output)
			}
		}
	}
	return banner + total
}
