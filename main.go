package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/MoneyprogerISG/smo-variant-20/smo"
)

func main() {
	n := flag.Int("n", 3, "число источников")
	m := flag.Int("m", 2, "число приборов")
	bufCap := flag.Int("buf", 3, "ёмкость буфера")
	a := flag.Float64("a", 1, "нижняя граница интервала между заявками")
	b := flag.Float64("b", 5, "верхняя граница интервала между заявками")
	lambda := flag.Float64("lambda", 0.5, "интенсивность обслуживания")
	seed := flag.Int64("seed", 1, "начальное значение генератора случайных чисел")
	flag.Parse()

	runStep(smo.Config{
		Sources:        *n,
		A:              *a,
		B:              *b,
		Devices:        *m,
		Lambda:         *lambda,
		BufferCapacity: *bufCap,
		Seed:           *seed,
	})
}

func runStep(cfg smo.Config) {
	sim := smo.NewSimulator(cfg)
	sim.Init()

	fmt.Printf("Источники: %d (τ от %g до %g) -> Буфер: %d мест -> Приборы: %d (λ=%g)\n",
		cfg.Sources, cfg.A, cfg.B, cfg.BufferCapacity, cfg.Devices, cfg.Lambda)
	fmt.Println("Enter — следующее событие, q — выход")

	printState(sim.GetState())

	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadString('\n')
		if strings.TrimSpace(line) == "q" || err != nil {
			return
		}

		e := sim.Step()
		if e == nil {
			fmt.Println("Календарь пуст")
			return
		}
		fmt.Printf("\n>>> %s\n", eventName(e))
		printState(sim.GetState())
	}
}

func printState(st smo.SystemState) {
	fmt.Printf("\nВремя: %.3f\n", st.Time)

	fmt.Println("Источники:")
	for _, s := range st.Sources {
		fmt.Printf("  И%d: сгенерировано %d, отказов %d (%.1f%%)\n",
			s.ID+1, s.Generated, s.Refused, s.RefusalProbability*100)
	}

	fmt.Println("Календарь:")
	for _, ev := range st.Calendar {
		fmt.Printf("  %s\n", eventName(ev))
	}

	fmt.Print("Буфер (голова слева): ")
	for i := 0; i < st.BufferCapacity; i++ {
		if i < len(st.Buffer) {
			fmt.Printf("[%s] ", requestName(st.Buffer[i]))
		} else {
			fmt.Print("[ ] ")
		}
	}
	fmt.Println()

	fmt.Println("Приборы:")
	for j, d := range st.Devices {
		mark := ""
		if j == st.Pointer {
			mark = "  <- указатель"
		}
		if d.Busy {
			fmt.Printf("  П%d: занят %s до %.3f%s\n", d.ID+1, requestName(d.Current), d.ReleaseTime, mark)
		} else {
			fmt.Printf("  П%d: свободен%s\n", d.ID+1, mark)
		}
	}
}

func eventName(e *smo.Event) string {
	switch e.Type() {
	case smo.GENERATION:
		return fmt.Sprintf("t=%.3f GENERATION И%d", e.Time(), e.Source().ID()+1)
	case smo.RELEASE:
		return fmt.Sprintf("t=%.3f RELEASE П%d", e.Time(), e.Device().ID()+1)
	default:
		return e.Type().String()
	}
}

func requestName(r *smo.Request) string {
	if r == nil {
		return "-"
	}
	return fmt.Sprintf("З%d.%d", r.SourceID()+1, r.Number())
}
