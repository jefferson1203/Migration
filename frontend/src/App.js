import React, { useState, useEffect } from 'react';
import SimulationCanvas from './components/SimulationCanvas';
import Controls from './components/Controls';
import Settings from './components/Settings';
import { fetchSimulationState, startSimulation, stopSimulation } from './utils/api';
import './styles/styles.css';

function App() {
  const [simulationState, setSimulationState] = useState({ birds: [], time: 0, isRunning: false, worldSize: 1000, obstacles: [], resources: [], collisionCount: 0 });
  const [isRunning, setIsRunning] = useState(false)

  
    useEffect(() => {
      const fetchState = async () => {
        try {
            const state = await fetchSimulationState();
            setSimulationState(state);
            setIsRunning(state.isRunning);
          } catch (error) {
            console.error("Failed to fetch simulation state:", error);
          }
      };
      fetchState();
  }, []);
  
    useEffect(() => {
        let intervalId;
        if (isRunning) {
              intervalId = setInterval(async () => {
                  try {
                          const state = await fetchSimulationState();
                          setSimulationState(state);
                      } catch (error) {
                        console.error("Failed to fetch simulation state:", error);
                      }
                }, 100) 
          }
        return () => clearInterval(intervalId)
    }, [isRunning]);


    const handleStart = async () => {
      try {
        await startSimulation();
        const state = await fetchSimulationState();
          setSimulationState(state);
        setIsRunning(state.isRunning);
        } catch (e) {
            console.log(e)
        }
  };

  const handleStop = async () => {
        try {
            await stopSimulation();
            const state = await fetchSimulationState();
             setSimulationState(state);
           setIsRunning(state.isRunning);
        } catch (e) {
            console.log(e)
        }
  };
    

  return (
    <div className="app-container">
      <h1>Bird Migration Simulation</h1>
      <div className="simulation-container">
        <SimulationCanvas simulationState={simulationState} />
        <div className="controls-container">
            <Controls isRunning={isRunning} onStart={handleStart} onStop={handleStop} />
            <Settings />
             <div>
                <h3>Collisions :</h3>
               <p>{simulationState.collisionCount}</p>
            </div>
        </div>
      </div>
    </div>
  );
}

export default App;