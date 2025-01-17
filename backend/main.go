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
	Port             int
	SimulationSpeed  int
	WorldSize        int
	InitialBirds     int
	EnvironmentSize  int
	DBPath           string
	ObstacleCount    int
	ResourceCount    int
	Temperature      float64
	FoodAvailability float64
	PredatorPresence float64
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

		config.Temperature, envErr = strconv.ParseFloat(getEnv("TEMPERATURE", "20.0"), 64)
		if envErr != nil {
			config.Temperature = 20.0
		}

		config.FoodAvailability, envErr = strconv.ParseFloat(getEnv("FOOD_AVAILABILITY", "1.0"), 64)
		if envErr != nil {
			config.FoodAvailability = 1.0
		}

		config.PredatorPresence, envErr = strconv.ParseFloat(getEnv("PREDATOR_PRESENCE", "0.0"), 64)
		if envErr != nil {
			config.PredatorPresence = 0.0
		}
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
	ID            int        `json:"id"`
	Position      [2]float64 `json:"position"`
	Velocity      [2]float64 `json:"velocity"`
	State         string     `json:"state"`
	Target        [2]float64 `json:"target"`
	Group         int        `json:"group"`
	CollisionTime int64      `json:"collisionTime"` // Time when the collision was first detected
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

type Predator struct {
	ID       int        `json:"id"`
	Position [2]float64 `json:"position"`
	Velocity [2]float64 `json:"velocity"`
}

type Zone struct {
	ID               int        `json:"id"`
	Position         [2]float64 `json:"position"`
	Temperature      float64    `json:"temperature"`
	FoodAvailability float64    `json:"foodAvailability"`
	PredatorPresence float64    `json:"predatorPresence"`
}

type SimulationState struct {
	Birds            []Bird            `json:"birds"`
	Time             int               `json:"time"`
	IsRunning        bool              `json:"isRunning"`
	WorldSize        int               `json:"worldSize"`
	Obstacles        []Obstacle        `json:"obstacles"`
	Resources        []Resource        `json:"resources"`
	CollisionCount   int               `json:"collisionCount"`
	Predators        []Predator        `json:"predators"`
	TemperatureZones []TemperatureZone `json:"temperatureZones"`
	Zones            []Zone            `json:"zones"`
}

type SimulationConfig struct {
	SimulationSpeed int `json:"simulationSpeed"`
	WorldSize       int `json:"worldSize"`
	InitialBirds    int `json:"initialBirds"`
	ObstacleCount   int `json:"obstacleCount"`
	ResourceCount   int `json:"resourceCount"`
}

type SaveState struct {
	State    SimulationState  `json:"state"`
	Config   SimulationConfig `json:"config"`
	TimeStep int              `json:"timeStep"`
}

type EnvironmentalFactors struct {
	Temperature      float64 `json:"temperature"`
	FoodAvailability float64 `json:"foodAvailability"`
	PredatorPresence float64 `json:"predatorPresence"`
}

type TemperatureZone struct {
	Region      int     `json:"region"`
	Temperature float64 `json:"temperature"`
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

// const minDistanceBetweenBirds = 100.0 // Increase minimum distance between birds when searching for food
const collisionThreshold = 2.0 // Distance threshold for collision detection
const separationDelay = 3000   // Delay in milliseconds before birds separate after finishing migration

var currentFoodLocation [2]float64
var foodRegion int

func detectCollisions() {
	for i := 0; i < len(simulationState.Birds); i++ {
		for j := i + 1; j < len(simulationState.Birds); j++ {
			bird1 := &simulationState.Birds[i]
			bird2 := &simulationState.Birds[j]
			dist := distance(bird1.Position, bird2.Position)
			if dist < collisionThreshold {
				if bird1.CollisionTime == 0 && bird2.CollisionTime == 0 {
					bird1.CollisionTime = int64(simulationState.Time)
					bird2.CollisionTime = int64(simulationState.Time)
					simulationState.CollisionCount++
					// Move birds apart to reduce further collisions
					moveBirdsApart(bird1, bird2)
				}
			} else {
				bird1.CollisionTime = 0
				bird2.CollisionTime = 0
			}
		}
	}
}

func moveBirdsApart(bird1, bird2 *Bird) {
	direction := [2]float64{bird1.Position[0] - bird2.Position[0], bird1.Position[1] - bird2.Position[1]}
	normalizedDirection := normalize(direction)
	bird1.Position[0] += normalizedDirection[0] * collisionThreshold
	bird1.Position[1] += normalizedDirection[1] * collisionThreshold
	bird2.Position[0] -= normalizedDirection[0] * collisionThreshold
	bird2.Position[1] -= normalizedDirection[1] * collisionThreshold
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

	router.POST("/environment", func(c *gin.Context) {
		var factors EnvironmentalFactors
		if err := c.ShouldBindJSON(&factors); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		SetEnvironmentalFactors(factors)
		c.JSON(http.StatusOK, gin.H{"message": "Environmental factors updated"})
	})

	router.GET("/environment", func(c *gin.Context) {
		factors := GetEnvironmentalFactors()
		c.JSON(http.StatusOK, factors)
	})

	router.GET("/temperature-zones", func(c *gin.Context) {
		c.JSON(http.StatusOK, simulationState.TemperatureZones)
	})

	router.POST("/temperature-zones", func(c *gin.Context) {
		var zones []TemperatureZone
		if err := c.ShouldBindJSON(&zones); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		simulationState.TemperatureZones = zones
		c.JSON(http.StatusOK, gin.H{"message": "Temperature zones updated"})
	})

	router.GET("/zones", func(c *gin.Context) {
		c.JSON(http.StatusOK, simulationState.Zones)
	})

	router.POST("/zones", func(c *gin.Context) {
		var zones []Zone
		if err := c.ShouldBindJSON(&zones); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		simulationState.Zones = zones
		c.JSON(http.StatusOK, gin.H{"message": "Zones updated"})
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

	// Ensure the number of resources is at least one-third of the number of birds
	resourceCount := config.ResourceCount
	if resourceCount < config.InitialBirds/3 {
		resourceCount = config.InitialBirds / 3
	}

	// Generate resources
	simulationState.Resources = make([]Resource, resourceCount) // Decrease the number of resources
	for i := range simulationState.Resources {
		resourceType := "food"
		if i%2 == 0 {
			resourceType = "rest"
		}
		simulationState.Resources[i] = Resource{
			ID:       i,
			Position: [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)},
			Type:     resourceType,
			Capacity: 5, // Decrease the capacity of resources
			Current:  5,
		}
	}

	numGroups := config.InitialBirds / 10
	if numGroups < 1 {
		numGroups = 1
	}
	groups := make([]int, config.InitialBirds)
	for i := range simulationState.Birds {
		groups[i] = rand.Intn(numGroups)
	}

	// Set one-third of the birds to "searchingFood" state and the rest to "migrating" state
	for i := range simulationState.Birds {
		state := "migrating"
		if i < config.InitialBirds/3 {
			state = "searchingFood"
		}
		simulationState.Birds[i] = Bird{
			ID:       i,
			Position: [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)},
			Velocity: [2]float64{rand.Float64() - 0.5, rand.Float64()*2 - 1},
			State:    state,
			Target:   [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)},
			Group:    groups[i],
		}
	}

	// Generate predators
	numPredators := int(config.PredatorPresence * 10) // Number of predators based on predator presence
	if numPredators < 1 {
		numPredators = 1 // Ensure at least one predator
	}
	simulationState.Predators = make([]Predator, numPredators)
	for i := range simulationState.Predators {
		simulationState.Predators[i] = Predator{
			ID:       i,
			Position: [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)},
			Velocity: [2]float64{rand.Float64() - 0.5, rand.Float64() - 0.5},
		}
	}

	// Generate initial food location on one side
	foodRegion = 0
	currentFoodLocation = generateFoodLocation(foodRegion)
	simulationState.Resources = []Resource{
		{
			ID:       0,
			Position: currentFoodLocation,
			Type:     "food",
			Capacity: 5,
			Current:  5,
		},
	}

	// Generate zones
	simulationState.Zones = []Zone{
		{ID: 0, Position: [2]float64{float64(config.WorldSize) / 4, float64(config.WorldSize) / 4}, Temperature: 15.0, FoodAvailability: 1.0, PredatorPresence: 0.1},
		{ID: 1, Position: [2]float64{3 * float64(config.WorldSize) / 4, float64(config.WorldSize) / 4}, Temperature: 25.0, FoodAvailability: 0.8, PredatorPresence: 0.2},
		{ID: 2, Position: [2]float64{float64(config.WorldSize) / 4, 3 * float64(config.WorldSize) / 4}, Temperature: 10.0, FoodAvailability: 0.5, PredatorPresence: 0.3},
		{ID: 3, Position: [2]float64{3 * float64(config.WorldSize) / 4, 3 * float64(config.WorldSize) / 4}, Temperature: 20.0, FoodAvailability: 0.9, PredatorPresence: 0.1},
	}

	// Generate initial food location in the best zone
	bestZone := findBestZone()
	currentFoodLocation = generateFoodLocation(bestZone.ID)
	simulationState.Resources = []Resource{
		{
			ID:       0,
			Position: currentFoodLocation,
			Type:     "food",
			Capacity: 5,
			Current:  5,
		},
	}

	simulationState.Time = 0
	simulationState.IsRunning = isRunning
	simulationState.WorldSize = config.WorldSize
	isRunning = true
}

func generateFoodLocation(region int) [2]float64 {
	var xOffset, yOffset float64
	switch region {
	case 0:
		xOffset, yOffset = 0, 0
	case 1:
		xOffset, yOffset = float64(config.WorldSize)/2, 0
	case 2:
		xOffset, yOffset = 0, float64(config.WorldSize)/2
	case 3:
		xOffset, yOffset = float64(config.WorldSize)/2, float64(config.WorldSize)/2
	}
	return [2]float64{xOffset + rand.Float64()*float64(config.WorldSize)/2, yOffset + rand.Float64()*float64(config.WorldSize)/2}
}

func findBestZone() Zone {
	bestZone := simulationState.Zones[0]
	for _, zone := range simulationState.Zones {
		if zone.Temperature > 10.0 && zone.Temperature < 25.0 && zone.FoodAvailability > bestZone.FoodAvailability {
			bestZone = zone
		}
	}
	return bestZone
}

func StartSimulation() {
	responseChan := make(chan bool)
	simulationControlChan <- simulationControlRequest{
		action:       "start",
		responseChan: responseChan,
	}
	result := <-responseChan
	log.Println(result)
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

	// Adjust bird behavior based on environmental factors
	for i := range simulationState.Birds {
		bird := &simulationState.Birds[i]
		zone := findClosestZone(bird.Position)
		if zone.Temperature < 10.0 {
			bird.State = "migrating"
		} else if zone.FoodAvailability < 0.5 {
			bird.State = "searchingFood"
		} else if zone.PredatorPresence > 0.5 {
			bird.State = "resting"
		}
	}

	groups := make(map[int][]Bird)
	for _, bird := range simulationState.Birds {
		groups[bird.Group] = append(groups[bird.Group], bird)
	}

	for group, birds := range groups {
		var totalX, totalY float64
		var numBirds int
		for _, bird := range birds {
			if bird.State == "migrating" {
				totalX += bird.Position[0]
				totalY += bird.Position[1]
				numBirds++
			}
		}
		var groupTarget [2]float64
		if numBirds > 0 {
			groupTarget = [2]float64{totalX / float64(numBirds), totalY / float64(numBirds)}
		} else {
			groupTarget = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
		}
		for i := range simulationState.Birds {
			bird := &simulationState.Birds[i]
			if bird.State == "migrating" && bird.Group == group {
				updateMigratingBird(i, groupTarget)
			} else if bird.State == "resting" {
				updateRestingBird(i)
			} else if bird.State == "searchingFood" {
				updateSearchingFoodBird(i)
			}
		}
	}

	// Check if the current food location is depleted
	if len(simulationState.Resources) == 0 || simulationState.Resources[0].Current <= 0 {
		// Generate new food location in the next best zone
		bestZone := findBestZone()
		currentFoodLocation = generateFoodLocation(bestZone.ID)
		simulationState.Resources = []Resource{
			{
				ID:       0,
				Position: currentFoodLocation,
				Type:     "food",
				Capacity: 5,
				Current:  5,
			},
		}

		// Update all birds to move towards the new food location
		for i := range simulationState.Birds {
			bird := &simulationState.Birds[i]
			bird.State = "migrating"
			bird.Target = currentFoodLocation
		}
	}

	// Limit the number of searching food birds
	searchingFoodBirds := 0
	for _, bird := range simulationState.Birds {
		if bird.State == "searchingFood" {
			searchingFoodBirds++
		}
	}

	// Randomly change states of birds
	for i := range simulationState.Birds {
		bird := &simulationState.Birds[i]
		if bird.State == "migrating" && rand.Float64() < 0.05 && searchingFoodBirds < len(simulationState.Resources)/4 {
			bird.State = "searchingFood"
			bird.Target = currentFoodLocation
			searchingFoodBirds++
		} else if bird.State == "searchingFood" && distance(bird.Position, bird.Target) < 10 {
			// After eating, change state to migrating or resting
			if rand.Float64() < 0.5 {
				bird.State = "migrating"
				bird.Target = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
			} else {
				bird.State = "resting"
			}
		} else if bird.State == "resting" && rand.Float64() < 0.1 {
			bird.State = "migrating"
			bird.Target = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
		}
	}

	// Update predator positions and check for attacks
	for i := range simulationState.Predators {
		predator := &simulationState.Predators[i]
		predator.Position[0] += predator.Velocity[0] * float64(timeStep)
		predator.Position[1] += predator.Velocity[1] * float64(timeStep)

		// Ensure predator stays within world boundaries
		predator.Position[0] = math.Max(0, math.Min(float64(config.WorldSize), predator.Position[0]))
		predator.Position[1] = math.Max(0, math.Min(float64(config.WorldSize), predator.Position[1]))

		// Check for attacks on birds
		for j := range simulationState.Birds {
			bird := &simulationState.Birds[j]
			if distance(predator.Position, bird.Position) < 10 {
				// Bird tries to escape by moving towards the nearest group
				closestGroupPos := findClosestGroup(bird)
				bird.Target = closestGroupPos
				bird.State = "migrating"
			}
		}
	}

	simulationState.Time++
}

func updateMigratingBird(i int, groupTarget [2]float64) {
	bird := &simulationState.Birds[i]

	// Move to target
	direction := [2]float64{groupTarget[0] - bird.Position[0], groupTarget[1] - bird.Position[1]}
	normalizedDirection := normalize(direction)
	bird.Velocity = [2]float64{normalizedDirection[0], normalizedDirection[1]}
	bird.Position[0] += bird.Velocity[0] * float64(timeStep)
	bird.Position[1] += bird.Velocity[1] * float64(timeStep)

	// Ensure bird stays within world boundaries
	bird.Position[0] = math.Max(0, math.Min(float64(config.WorldSize), bird.Position[0]))
	bird.Position[1] = math.Max(0, math.Min(float64(config.WorldSize), bird.Position[1]))

	// Evade obstacles
	evadeObstacles(i)

	// Change state based on time and proximity to resources
	if simulationState.Time%500 == 0 { // Resting state change
		closestResource, _ := findClosestResource(bird.Position, "rest")
		if closestResource != nil && distance(bird.Position, closestResource.Position) < 50 {
			bird.State = "resting"
			closestResource.Current++
		}
	} else if simulationState.Time%300 == 0 { // Searching food state change
		closestResource, _ := findClosestResource(bird.Position, "food")
		if closestResource != nil && closestResource.Current < closestResource.Capacity && distance(bird.Position, closestResource.Position) < 50 {
			bird.State = "searchingFood"
			bird.Target = closestResource.Position
			closestResource.Current++
		}
	}
}

func updateRestingBird(i int) {
	bird := &simulationState.Birds[i]
	// Change state after resting
	if simulationState.Time%(separationDelay) == 0 {
		bird.State = "migrating"
		closestResource, index := findClosestResource(bird.Position, "rest")
		if closestResource != nil {
			simulationState.Resources[index].Current--
		}
		// Set a new target within a small circle
		angle := rand.Float64() * 2 * math.Pi
		radius := rand.Float64() * 10
		bird.Target = [2]float64{
			bird.Position[0] + radius*math.Cos(angle),
			bird.Position[1] + radius*math.Sin(angle),
		}
	}
}

func updateSearchingFoodBird(i int) {
	bird := &simulationState.Birds[i]
	closestResource, index := findClosestResource(bird.Position, "food")
	if closestResource == nil || closestResource.Current <= 0 {
		// If no food is available nearby, move to a random location to search for food
		bird.Target = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
		return
	}
	// Move to target
	direction := [2]float64{closestResource.Position[0] - bird.Position[0], closestResource.Position[1] - bird.Position[1]}
	normalizedDirection := normalize(direction)
	bird.Velocity = [2]float64{normalizedDirection[0], normalizedDirection[1]}
	bird.Position[0] += bird.Velocity[0] * float64(timeStep)
	bird.Position[1] += bird.Velocity[1] * float64(timeStep)

	// Ensure bird stays within world boundaries
	bird.Position[0] = math.Max(0, math.Min(float64(config.WorldSize), bird.Position[0]))
	bird.Position[1] = math.Max(0, math.Min(float64(config.WorldSize), bird.Position[1]))

	if distance(bird.Position, closestResource.Position) < 10 {
		bird.State = "migrating"
		simulationState.Resources[index].Current--
		if simulationState.Resources[index].Current <= 0 {
			// Move the resource to a new location if it is depleted
			simulationState.Resources[index].Position = [2]float64{rand.Float64() * float64(config.WorldSize), rand.Float64() * float64(config.WorldSize)}
			simulationState.Resources[index].Current = simulationState.Resources[index].Capacity
		}
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
	dy := pos2[1] - pos2[1]
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

func findClosestGroup(bird *Bird) [2]float64 {
	groups := make(map[int][]Bird)
	for _, b := range simulationState.Birds {
		groups[b.Group] = append(groups[b.Group], b)
	}

	var closestGroupPos [2]float64
	minDist := math.MaxFloat64
	for _, group := range groups {
		if len(group) > 1 {
			var totalX, totalY float64
			for _, b := range group {
				totalX += b.Position[0]
				totalY += b.Position[1]
			}
			groupPos := [2]float64{totalX / float64(len(group)), totalY / float64(len(group))}
			dist := distance(bird.Position, groupPos)
			if dist < minDist {
				minDist = dist
				closestGroupPos = groupPos
			}
		}
	}
	return closestGroupPos
}

func findClosestZone(pos [2]float64) Zone {
	var closest Zone
	minDist := math.MaxFloat64
	for _, zone := range simulationState.Zones {
		dist := distance(pos, zone.Position)
		if dist < minDist {
			minDist = dist
			closest = zone
		}
	}
	return closest
}

func startSimulationLoop() {
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
				InitialBirds:    config.InitialBirds,
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
				req.responseChan <- true

			case "stop":
				isRunning = false
				simulationState.IsRunning = isRunning
				req.responseChan <- true
			case "restart":
				initSimulation()
				isRunning = true
				simulationState.IsRunning = isRunning
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
	config.InitialBirds = newConfig.InitialBirds
	config.ObstacleCount = newConfig.ObstacleCount
	config.ResourceCount = newConfig.ResourceCount

	simulationState.WorldSize = config.WorldSize
	simulationState.Birds = make([]Bird, config.InitialBirds)
	initSimulation()
}

func SetTimeStep(newTimeStep int) {
	responseChan := make(chan int)
	timeStepChan <- timeStepRequest{
		newTimeStep:  newTimeStep,
		responseChan: responseChan,
	}
	newTimeStepValue := <-responseChan
	timeStep = newTimeStepValue
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
		InitialBirds:    config.InitialBirds,
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
	config.InitialBirds = loadedConfig.InitialBirds
	config.ObstacleCount = loadedConfig.ObstacleCount
	config.ResourceCount = loadedConfig.ResourceCount
	simulationState.IsRunning = false
	isRunning = false

	return &SaveState{
		State:    loadedState,
		Config:   loadedConfig,
		TimeStep: timeStep,
	}, nil
}

func SetEnvironmentalFactors(factors EnvironmentalFactors) {
	config.Temperature = factors.Temperature
	config.FoodAvailability = factors.FoodAvailability
	config.PredatorPresence = factors.PredatorPresence

	// Reinitialize simulation to update the number of predators
	initSimulation()
}

func GetEnvironmentalFactors() EnvironmentalFactors {
	return EnvironmentalFactors{
		Temperature:      config.Temperature,
		FoodAvailability: config.FoodAvailability,
		PredatorPresence: config.PredatorPresence,
	}
}
