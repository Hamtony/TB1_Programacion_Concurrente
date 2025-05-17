#define NUM_SAMPLES 20
#define NUM_WORKERS 4
#define INPUT_SIZE 5

// Semáforo binario
byte mutex = 1

// Simulación de entradas y salidas (sin usar arrays 2D)
byte input_i[NUM_SAMPLES * INPUT_SIZE] // entradas simuladas 1D
byte outputs[NUM_SAMPLES]              // resultados de "predicción"

// Funciones para controlar concurrencia
inline wait(s) {
    do
    :: s > 0 -> s--
    :: else -> skip
    od
}

inline signal(s) {
    s++
}

// Proceso principal
init {
    byte i, j;

    // Simula inicialización de entradas como si fuera 2D
    i = 0;
    do
    :: i < NUM_SAMPLES ->
        j = 0;
        do
        :: j < INPUT_SIZE ->
            input_i[i * INPUT_SIZE + j] = (i + j) % 10;
            j++
        :: else -> break
        od;
        i++
    :: else -> break
    od;

    // Lanzar procesos concurrentes
    atomic {
        run worker(0);
        run worker(1);
        run worker(2);
        run worker(3);
    }
}

// Proceso concurrente simulando feedforward
proctype worker(byte id) {
    byte start = (id * NUM_SAMPLES) / NUM_WORKERS;
    byte end = ((id + 1) * NUM_SAMPLES) / NUM_WORKERS;
    byte i;

    i = start;
    do
    :: i < end ->

        // Sección crítica
        wait(mutex);
        outputs[i] = 1; // Simulación de valor de predicción
        assert(outputs[i] == 1); // Validación del valor
        signal(mutex);

        i++
    :: else -> break
    od;
}
