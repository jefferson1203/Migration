import React, { useState, useEffect, useCallback } from 'react';
import SimulationCanvas from './components/SimulationCanvas';
import Controls from './components/Controls';
import Settings from './components/Settings';
import { fetchSimulationState, startSimulation, stopSimulation } from './utils/api';
import './styles/styles.css';

function App() {
  const [simulationState, setSimulationState] = useState({
    birds: [],
    time: 0,
    isRunning: false,
    worldSize: 1000,
    obstacles: [],
    resources: [],
    collisionCount: 0,
  });

  // Fetch the simulation state on component mount
  useEffect(() => {
    const fetchInitialState = async () => {
      try {
        const state = await fetchSimulationState();
        setSimulationState((prevState) => ({ ...prevState, ...state }));
      } catch (error) {
        console.error('Failed to fetch initial simulation state:', error);
      }
    };
    fetchInitialState();
  }, []);

  // Periodically update simulation state when running
  useEffect(() => {
    if (simulationState.isRunning) {
      const intervalId = setInterval(async () => {
        try {
          const state = await fetchSimulationState();
          setSimulationState((prevState) => ({ ...prevState, ...state }));
        } catch (error) {
          console.error('Failed to fetch simulation state:', error);
        }
      }, 100);

      return () => clearInterval(intervalId);
    }
  }, [simulationState.isRunning]);

  // Start simulation handler
  const handleStart = useCallback(async () => {
    try {
      await startSimulation();
      const state = await fetchSimulationState();
      setSimulationState((prevState) => ({ ...prevState, ...state }));
    } catch (error) {
      console.error('Failed to start simulation:', error);
    }
  }, []);

  // Stop simulation handler
  const handleStop = useCallback(async () => {
    try {
      await stopSimulation();
      const state = await fetchSimulationState();
      setSimulationState((prevState) => ({ ...prevState, ...state }));
    } catch (error) {
      console.error('Failed to stop simulation:', error);
    }
  }, []);

  return (
    <div className="app-container">
      <h1>Bird Migration Simulation</h1>
      <div className="simulation-container">
        {/* Simulation canvas */}
        <SimulationCanvas simulationState={simulationState} />

        {/* Controls and other information */}
        <div className="controls-container">
          <Controls
            isRunning={simulationState.isRunning}
            onStart={handleStart}
            onStop={handleStop}
          />
          <Settings />
          <div>
            <h3>Collisions:</h3>
            <p>{simulationState.collisionCount}</p>
          </div>
          <div>
            <h3>Legend:</h3>
            <p><span style={{ color: 'green' }}>●</span> Migrating Bird</p>
            <p><span style={{ color: 'blue' }}>●</span> Resting Bird</p>
            <p><span style={{ color: 'purple' }}>●</span> Searching Food Bird</p>
            <p><span style={{ color: 'orange' }}>●</span> Food Resource</p>
            <p><span style={{ color: 'lightblue' }}>●</span> Rest Resource</p>
            <p><span style={{ color: 'gray' }}>●</span> Obstacle</p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;