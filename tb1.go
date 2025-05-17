package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type ConcurrentNeuralNetwork struct {
	weights []float64
	bias    float64
}

func sigmoidConcurrent(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func NewConcurrentNeuralNetwork(inputSize int) *ConcurrentNeuralNetwork {
	weights := make([]float64, inputSize)
	for i := range weights {
		weights[i] = rand.Float64()*2 - 1
	}
	bias := rand.Float64()*2 - 1
	return &ConcurrentNeuralNetwork{weights: weights, bias: bias}
}

func (nn *ConcurrentNeuralNetwork) PredictConcurrent(input []float64) float64 {
	sum := 0.0
	for i := range input {
		sum += input[i] * nn.weights[i]
	}
	sum += nn.bias
	return sigmoidConcurrent(sum)
}

func (nn *ConcurrentNeuralNetwork) PredictBatchConcurrent(inputs [][]float64, numWorkers int) []float64 {
	numSamples := len(inputs)
	outputs := make([]float64, numSamples)
	var wg sync.WaitGroup

	blockSize := (numSamples + numWorkers - 1) / numWorkers

	for worker := 0; worker < numWorkers; worker++ {
		start := worker * blockSize
		end := (worker + 1) * blockSize
		if end > numSamples {
			end = numSamples
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				outputs[i] = nn.PredictConcurrent(inputs[i])
			}
		}(start, end)
	}

	wg.Wait()
	return outputs
}

// Preprocesamiento de CSV y codificación simple
func loadCSVData(filename string) ([][]float64, []float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(records) < 1 {
		return nil, nil, fmt.Errorf("el archivo está vacío")
	}

	header := records[0]
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	// Mapas para codificar strings
	regionMap := map[string]int{}
	manufacturerMap := map[string]int{}
	conditionMap := map[string]int{}

	var inputs [][]float64
	var targets []float64

	for _, row := range records[1:] {
		region := row[colIndex["region"]]
		priceStr := row[colIndex["price"]]
		yearStr := row[colIndex["year"]]
		manufacturer := row[colIndex["manufacturer"]]
		condition := row[colIndex["condition"]]
		odometerStr := row[colIndex["odometer"]]

		if priceStr == "" || yearStr == "" || odometerStr == "" {
			continue
		}

		price, err1 := strconv.ParseFloat(priceStr, 64)
		year, err2 := strconv.ParseFloat(yearStr, 64)
		odometer, err3 := strconv.ParseFloat(odometerStr, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}

		// Codificar categóricos
		if _, ok := regionMap[region]; !ok {
			regionMap[region] = len(regionMap)
		}
		if _, ok := manufacturerMap[manufacturer]; !ok {
			manufacturerMap[manufacturer] = len(manufacturerMap)
		}
		if _, ok := conditionMap[condition]; !ok {
			conditionMap[condition] = len(conditionMap)
		}

		regionCode := float64(regionMap[region])
		manufacturerCode := float64(manufacturerMap[manufacturer])
		conditionCode := float64(conditionMap[condition])

		input := []float64{
			regionCode,     // categórico mapeado a número
			year,           // año
			manufacturerCode,
			conditionCode,
			odometer,
		}

		inputs = append(inputs, input)
		targets = append(targets, price)
	}

	return inputs, targets, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Leyendo archivo CSV...")
	inputs, targets, err := loadCSVData("vehicles.csv")
	if err != nil {
		log.Fatal("Error al leer CSV:", err)
	}

	if len(inputs) == 0 {
		log.Fatal("No se cargaron datos del archivo CSV.")
	}

	inputSize := len(inputs[0])
	numWorkers := 8

	nn := NewConcurrentNeuralNetwork(inputSize)

	start := time.Now()
	preds := nn.PredictBatchConcurrent(inputs, numWorkers)
	duration := time.Since(start)

	fmt.Printf("Concurrente: procesados %d registros en %v\n", len(preds), duration)

	// Mostrar algunas predicciones
	for i := 0; i < 5 && i < len(preds); i++ {
		fmt.Printf("Real: $%.2f | Predicho: %.2f\n", targets[i], preds[i])
	}
}
