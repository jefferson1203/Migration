package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// --- Configuration ---
var config struct {
	Port            int
	SimulationSpeed int
	WorldSize       int
	InitialBirds    int
	EnvironmentSize int
	DBPath          string
	ObstacleCount   int
	ResourceCount   int
}

var once sync.Once

func LoadConfig() {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file")
		}
		var envErr error
		config.Port, envErr = strconv.Atoi(getEnv("PORT", "8080"))
		if envErr != nil {
			config.Port = 8080
		}

		config.SimulationSpeed, envErr = strconv.Atoi(getEnv("SIMULATION_SPEED", "100"))
		if envErr != nil {
			config.SimulationSpeed = 100
		}

		config.WorldSize, envErr = strconv.Atoi(getEnv("WORLD_SIZE", "1000"))
		if envErr != nil {
			config.WorldSize = 1000
		}

		config.InitialBirds, envErr = strconv.Atoi(getEnv("INITIAL_BIRDS", "50"))
		if envErr != nil {
			config.InitialBirds = 50
		}

		config.EnvironmentSize, envErr = strconv.Atoi(getEnv("ENVIRONMENT_SIZE", "100"))
		if envErr != nil {
			config.EnvironmentSize = 100
		}

		config.ObstacleCount, envErr = strconv.Atoi(getEnv("OBSTACLE_COUNT", "5"))
		if envErr != nil {
			config.ObstacleCount = 5
		}

		config.ResourceCount, envErr = strconv.Atoi(getEnv("RESOURCE_COUNT", "5"))
		if envErr != nil {
			config.ResourceCount = 5
		}

		config.DBPath = getEnv("DB_PATH", "simulation.db")
		log.Println("config loaded", config)
	})
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// --- Models ---
type Bird struct {
	ID       int        `json:"id"`
	Position [2]float64 `json:"position"`
	Velocity [2]float64 `json:"velocity"`
	State    string     `json:"state"`
	Target   [2]float64 `json:"target"`
}

type Obstacle struct {
	ID       int        `json:"id"`
	Position [2]float64 `json:"position"`
	Radius   float64    `json:"radius"`
}

type Resource struct {
	ID       int        `json:"id"`
	Position [2]float64 `json:"position"`
	Type     string     `json:"type"`
	Capacity int        `json:"capacity"`
	Current  int        `json:"current"`
}

type SimulationState struct {
	Birds          []Bird     `json:"birds"`
	Time           int        `json:"time"`
	IsRunning      bool       `json:"isRunning"`
	WorldSize      int        `json:"worldSize"`
	Obstacles      []Obstacle `json:"obstacles"`
	Resources      []Resource `json:"resources"`
	CollisionCount int        `json:"collisionCount"`
}

type SimulationConfig struct {
	SimulationSpeed int `json:"simulationSpeed"`
	WorldSize       int `json:"worldSize"`
	ObstacleCount   int `json:"obstacleCount"`
	ResourceCount   int `json:"resourceCount"`
}

type SaveState struct {
	State    SimulationState  `json:"state"`
	Config   SimulationConfig `json:"config"`
	TimeStep int              `json:"timeStep"`
}

// --- Simulation Engine ---
var (
	simulationState SimulationState
	isRunning       bool
	timeStep        int = 1
	db              *sql.DB

	// Channels for synchronisation
	stateChan             chan simulationRequest
	configChan            chan configRequest
	timeStepChan          chan timeStepRequest
	simulationControlChan chan simulationControlRequest
)

type simulationRequest struct {
	responseChan chan SimulationState
}

type configRequest struct {
	responseChan chan SimulationConfig
}

type timeStepRequest struct {
	newTimeStep  int
	responseChan chan int
}

type simulationControlRequest struct {
	action       string
	responseChan chan bool
}

func detectCollisions() {
	for i := 0; i < len(simulationState.Birds); i++ {
		for j := i + 1; j < len(simulationState.Birds); j++ {
			bird1 := &simulationState.Birds[i]
			bird2 := &simulationState.Birds[j]
			dist := distance(bird1.Position, bird2.Position)
			if dist < 2 {
				simulationState.CollisionCount++
				log.Printf("Collision detected between bird %d and bird %d, current count: %d \n", bird1.ID, bird2.ID, simulationState.CollisionCount)
			}
		}
	}
}

func main() {
	LoadConfig()

	// Initialize SQLite database
	var err error
	db, err = sql.Open("sqlite3", config.DBPath)
	if err != nil {
		log.Fatal(err)
	}

	err = initDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Init simulation
	initSimulation()

	//init channels
	stateChan = make(chan simulationRequest)
	configChan = make(chan configRequest)
	timeStepChan = make(chan timeStepRequest)
	simulationControlChan = make(chan simulationControlRequest)

	go startSimulationLoop()

	router := gin.Default()

	// Set up CORS middleware
	Corsconfig := cors.DefaultConfig()
	Corsconfig.AllowAllOrigins = true
	Corsconfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	Corsconfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	router.Use(cors.New(Corsconfig))

	// Define endpoints directly in main.go
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.POST("/simulation/start", func(c *gin.Context) {
		StartSimulation()
		c.JSON(http.StatusOK, gin.H{"message": "Simulation started"})
	})

	router.POST("/simulation/stop", func(c *gin.Context) {
		StopSimulation()
		c.JSON(http.StatusOK, gin.H{"message": "Simulation stopped"})
	})

	router.GET("/simulation", func(c *gin.Context) {
		log.Println(" GET /simulation called")
		state := GetSimulationState()
		c.JSON(http.StatusOK, state)
	})

	router.GET("/simulation/config", func(c *gin.Context) {
		config := GetSimulationConfig()
		c.JSON(http.StatusOK, config)
	})

	router.POST("/simulation/config", func(c *gin.Context) {
		var newConfig SimulationConfig
		if err := c.ShouldBindJSON(&newConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		SetSimulationConfig(newConfig)
		c.JSON(http.StatusOK, gin.H{"message": "Simulation config updated"})
	})

	router.POST("/simulation/time-step", func(c *gin.Context) {
		var newTimeStep struct {
			TimeStep int `json:"timeStep"`
		}
		if err := c.ShouldBindJSON(&newTimeStep); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		SetTimeStep(newTimeStep.TimeStep)
		c.JSON(http.StatusOK, gin.H{"message": "Time step updated"})
	})

	router.GET("/simulation/time-step", func(c *gin.Context) {
		step := GetTimeStep()
		c.JSON(http.StatusOK, gin.H{"timeStep": step})
	})

	router.POST("/simulation/save", func(c *gin.Context) {
		err := SaveSimulationState()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Simulation state saved"})
	})

	router.GET("/simulation/load", func(c *gin.Context) {
		savedState, err := LoadSimulationState()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, savedState)
	})

	fmt.Printf("Server running on http://localhost:%d\n", config.Port)
	if err := router.Run(fmt.Sprintf(":%d", config.Port)); err != nil {
		log.Fatal(err)
	}
}

func initDatabase() error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS saved_states (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            state TEXT,
			config TEXT,
			time_step INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return fmt.Errorf("error creating saved_states table: %w", err)
	}
	return nil
}

func initSimulation() {
	simulationState.Birds = make([]Bird, config.InitialBirds)
	rand.Seed(time.Now().UnixNano())

	// Generate obstacles
	simulationState.Obstacles = make([]Obstacle, config.ObstacleCount)
	for i := range simulationState.Obstacles {
		simulationState.Obstacles[i] = Obstacle{
			ID:       i,
			Position: [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)},
			Radius:   rand.Float64()*15 + 5,
		}
	}

	// Generate resources
	simulationState.Resources = make([]Resource, config.ResourceCount)
	for i := range simulationState.Resources {
		resourceType := "food"
		if i%2 == 0 {
			resourceType = "rest"
		}
		simulationState.Resources[i] = Resource{
			ID:       i,
			Position: [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)},
			Type:     resourceType,
			Capacity: 10,
			Current:  0,
		}
	}

	for i := range simulationState.Birds {
		simulationState.Birds[i] = Bird{
			ID:       i,
			Position: [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)},
			Velocity: [2]float64{rand.Float64() - 0.5, rand.Float64()*2 - 1},
			State:    "migrating",
			Target:   [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)},
		}
	}
	simulationState.Time = 0
	simulationState.IsRunning = isRunning
	simulationState.WorldSize = config.WorldSize
	isRunning = true
	simulationState.IsRunning = isRunning
	log.Println("simulation state initialized ", simulationState)
}

func StartSimulation() {
	log.Println("StartSimulation called")
	responseChan := make(chan bool)
	simulationControlChan <- simulationControlRequest{
		action:       "start",
		responseChan: responseChan,
	}
	result := <-responseChan
	log.Println("responseChan return: ", result)
}

func StopSimulation() {
	responseChan := make(chan bool)
	simulationControlChan <- simulationControlRequest{
		action:       "stop",
		responseChan: responseChan,
	}
	<-responseChan
}

func updateSimulation() {
	if !isRunning {
		return
	}

	for i := range simulationState.Birds {
		bird := &simulationState.Birds[i]
		log.Printf("bird %d, state: %s before update, position %f, %f", bird.ID, bird.State, bird.Position[0], bird.Position[1])
		switch bird.State {
		case "migrating":
			updateMigratingBird(i)
		case "resting":
			updateRestingBird(i)
		case "searchingFood":
			updateSearchingFoodBird(i)
		}
		log.Printf("bird %d, state: %s after update, position %f, %f", bird.ID, bird.State, bird.Position[0], bird.Position[1])
	}
	simulationState.Time++
}

func updateMigratingBird(i int) {
	bird := &simulationState.Birds[i]
	//find next target
	if distance(bird.Position, bird.Target) < 10 {
		bird.Target = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
	}

	//move to target
	direction := [2]float64{bird.Target[0] - bird.Position[0], bird.Target[1] - bird.Position[1]}
	normalizedDirection := normalize(direction)
	bird.Velocity = [2]float64{normalizedDirection[0], normalizedDirection[1]}
	bird.Position[0] += bird.Velocity[0] * float64(timeStep)
	bird.Position[1] += bird.Velocity[1] * float64(timeStep)

	// Evade obstacles
	evadeObstacles(i)

	// Simple rule: birds rest after migrating for a while, or when they are near a rest resource
	if simulationState.Time%(1000) == 0 {
		closestResource, _ := findClosestResource(bird.Position, "rest")
		if closestResource != nil && distance(bird.Position, closestResource.Position) < 50 {
			bird.State = "resting"
			closestResource.Current++
		}
	} else if simulationState.Time%(700) == 0 {
		closestResource, _ := findClosestResource(bird.Position, "food")
		if closestResource != nil && closestResource.Current < closestResource.Capacity && distance(bird.Position, closestResource.Position) < 50 {
			bird.State = "searchingFood"
			closestResource.Current++
		}
	}
}

func updateRestingBird(i int) {
	bird := &simulationState.Birds[i]
	// after resting, resume migration
	if simulationState.Time%(1000) == 100 {
		bird.State = "migrating"
		closestResource, index := findClosestResource(bird.Position, "rest")
		if closestResource != nil {
			simulationState.Resources[index].Current--
		}
		bird.Target = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
	}
}

func updateSearchingFoodBird(i int) {
	bird := &simulationState.Birds[i]
	closestResource, index := findClosestResource(bird.Position, "food")
	if closestResource == nil || closestResource.Current >= closestResource.Capacity {
		bird.State = "migrating"
		bird.Target = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
		return
	}
	//move to target
	direction := [2]float64{closestResource.Position[0] - bird.Position[0], closestResource.Position[1] - bird.Position[1]}
	normalizedDirection := normalize(direction)
	bird.Velocity = [2]float64{normalizedDirection[0], normalizedDirection[1]}
	bird.Position[0] += bird.Velocity[0] * float64(timeStep)
	bird.Position[1] += bird.Velocity[1] * float64(timeStep)
	if distance(bird.Position, closestResource.Position) < 10 {
		bird.State = "migrating"
		simulationState.Resources[index].Current--
		bird.Target = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
	}
}

func evadeObstacles(i int) {
	bird := &simulationState.Birds[i]
	for _, obstacle := range simulationState.Obstacles {
		dist := distance(bird.Position, obstacle.Position)
		if dist < obstacle.Radius+10 {
			evadeDirection := [2]float64{bird.Position[0] - obstacle.Position[0], bird.Position[1] - obstacle.Position[1]}
			normalizedEvade := normalize(evadeDirection)
			bird.Velocity = [2]float64{normalizedEvade[0] * 0.5, normalizedEvade[1] * 0.5}
			return
		}
	}
}

func distance(pos1 [2]float64, pos2 [2]float64) float64 {
	dx := pos1[0] - pos2[0]
	dy := pos1[1] - pos2[1]
	return math.Sqrt(dx*dx + dy*dy)
}

func normalize(vec [2]float64) [2]float64 {
	mag := math.Sqrt(vec[0]*vec[0] + vec[1]*vec[1])
	if mag > 0 {
		return [2]float64{vec[0] / mag, vec[1] / mag}
	}
	return vec
}

func findClosestResource(pos [2]float64, resourceType string) (*Resource, int) {
	var closest *Resource
	var closestIndex int
	minDist := math.MaxFloat64
	found := false

	for index, res := range simulationState.Resources {
		if res.Type == resourceType {
			dist := distance(pos, res.Position)
			if dist < minDist {
				minDist = dist
				closest = &res
				closestIndex = index
				found = true
			}
		}
	}
	if found {
		return closest, closestIndex
	}

	return nil, -1
}

func startSimulationLoop() {
	log.Println("startSimulationLoop started")
	ticker := time.NewTicker(time.Duration(config.SimulationSpeed) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			updateSimulation()
			detectCollisions()
		case req := <-stateChan:
			req.responseChan <- simulationState
		case req := <-configChan:
			req.responseChan <- SimulationConfig{
				SimulationSpeed: config.SimulationSpeed,
				WorldSize:       config.WorldSize,
				ObstacleCount:   config.ObstacleCount,
				ResourceCount:   config.ResourceCount,
			}
		case req := <-timeStepChan:
			timeStep = req.newTimeStep
			req.responseChan <- timeStep
		case req := <-simulationControlChan:
			switch req.action {
			case "start":
				isRunning = true
				simulationState.IsRunning = isRunning
				log.Println("Simulation start requested, current isRunning: ", isRunning)
				req.responseChan <- true

			case "stop":
				isRunning = false
				simulationState.IsRunning = isRunning
				log.Println("Simulation stop requested, current isRunning: ", isRunning)
				req.responseChan <- true
			case "restart":
				initSimulation()
				isRunning = true
				simulationState.IsRunning = isRunning
				log.Println("Simulation restart requested, current isRunning: ", isRunning)
				req.responseChan <- true
			}
		}
	}
}

func GetSimulationState() SimulationState {
	responseChan := make(chan SimulationState)
	stateChan <- simulationRequest{
		responseChan: responseChan,
	}
	state := <-responseChan
	return state
}

func GetSimulationConfig() SimulationConfig {
	responseChan := make(chan SimulationConfig)
	configChan <- configRequest{
		responseChan: responseChan,
	}
	return <-responseChan
}

func SetSimulationConfig(newConfig SimulationConfig) {
	responseChan := make(chan SimulationConfig)
	configChan <- configRequest{
		responseChan: responseChan,
	}
	setSimulationConfigHelper(newConfig)
	<-responseChan
}

func setSimulationConfigHelper(newConfig SimulationConfig) {

	config.SimulationSpeed = newConfig.SimulationSpeed
	config.WorldSize = newConfig.WorldSize
	config.ObstacleCount = newConfig.ObstacleCount
	config.ResourceCount = newConfig.ResourceCount

	simulationState.WorldSize = config.WorldSize

}

func SetTimeStep(newTimeStep int) {
	responseChan := make(chan int)
	timeStepChan <- timeStepRequest{
		newTimeStep:  newTimeStep,
		responseChan: responseChan,
	}
	newTimeStepValue := <-responseChan
	timeStep = newTimeStepValue
	log.Printf("Time step set to %d \n", timeStep)
}

func GetTimeStep() int {
	responseChan := make(chan int)
	timeStepChan <- timeStepRequest{
		responseChan: responseChan,
	}
	return <-responseChan
}

func SaveSimulationState() error {

	stateJSON, err := json.Marshal(simulationState)
	if err != nil {
		return fmt.Errorf("error marshaling simulation state: %w", err)
	}

	configJSON, err := json.Marshal(SimulationConfig{
		SimulationSpeed: config.SimulationSpeed,
		WorldSize:       config.WorldSize,
		ObstacleCount:   config.ObstacleCount,
		ResourceCount:   config.ResourceCount,
	})
	if err != nil {
		return fmt.Errorf("error marshaling simulation config: %w", err)
	}

	_, err = db.Exec("INSERT INTO saved_states (state, config, time_step) VALUES (?, ?, ?)", stateJSON, configJSON, timeStep)
	if err != nil {
		return fmt.Errorf("error saving simulation state to DB: %w", err)
	}

	return nil
}

func LoadSimulationState() (*SaveState, error) {
	var stateJSON string
	var configJSON string
	var timeStep int

	row := db.QueryRow("SELECT state, config, time_step FROM saved_states ORDER BY id DESC LIMIT 1")
	err := row.Scan(&stateJSON, &configJSON, &timeStep)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no saved state found")
	}
	if err != nil {
		return nil, fmt.Errorf("error loading simulation state from DB: %w", err)
	}

	var loadedState SimulationState
	if err := json.Unmarshal([]byte(stateJSON), &loadedState); err != nil {
		return nil, fmt.Errorf("error unmarshaling simulation state: %w", err)
	}

	var loadedConfig SimulationConfig
	if err := json.Unmarshal([]byte(configJSON), &loadedConfig); err != nil {
		return nil, fmt.Errorf("error unmarshaling simulation config: %w", err)
	}

	simulationState = loadedState
	config.SimulationSpeed = loadedConfig.SimulationSpeed
	config.WorldSize = loadedConfig.WorldSize
	config.ObstacleCount = loadedConfig.ObstacleCount
	config.ResourceCount = loadedConfig.ResourceCount
	simulationState.IsRunning = false
	isRunning = false
	timeStep = timeStep

	return &SaveState{
		State:    loadedState,
		Config:   loadedConfig,
		TimeStep: timeStep,
	}, nil

}
