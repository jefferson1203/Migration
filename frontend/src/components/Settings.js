import React, { useState, useEffect } from 'react';
import { fetchSimulationConfig, updateSimulationConfig, setTimeStep, fetchTimeStep } from '../utils/api';

const Settings = () => {
  const [config, setConfig] = useState({ simulationSpeed: 100, worldSize: 1000, initialBirds: 50 });
  const [timeStep, setTimeStepState] = useState(1);


  useEffect(() => {
    const loadConfig = async () => {
      try {
        const fetchedConfig = await fetchSimulationConfig();
        setConfig(fetchedConfig);

      } catch (error) {
        console.error("Failed to load config:", error)
      }
    }
    
    const loadTimeStep = async () => {
      try {
        const fetchedTimeStep = await fetchTimeStep();
        setTimeStepState(fetchedTimeStep.timeStep);
      } catch (error) {
          console.error("Failed to load time step:", error)
      }
    }

    loadConfig()
    loadTimeStep()
  }, []);


  const handleConfigChange = async (e) => {
    const { name, value } = e.target;
    setConfig(prevConfig => ({ ...prevConfig, [name]: parseInt(value, 10) }));
    try {
        await updateSimulationConfig({
            ...config,
            [name]: parseInt(value, 10),
          });
    } catch (error) {
       console.error("Failed to update config:", error)
    }
  };


  const handleTimeStepChange = async (e) => {
    const value = e.target.value
    setTimeStepState(parseInt(value, 10));
    try {
      await setTimeStep(parseInt(value,10));
    } catch (error) {
        console.error("Failed to set time step:", error)
    }
  };


  return (
    <div>
        <h3>Settings</h3>
        <div>
          <label>Simulation Speed (ms): </label>
          <input
            type="number"
            name="simulationSpeed"
            value={config.simulationSpeed}
            onChange={handleConfigChange}
          />
        </div>

         <div>
          <label>World Size: </label>
          <input
            type="number"
            name="worldSize"
            value={config.worldSize}
            onChange={handleConfigChange}
          />
        </div>
         <div>
          <label>Initial Birds: </label>
          <input
            type="number"
            name="initialBirds"
            value={config.initialBirds}
            onChange={handleConfigChange}
          />
        </div>
         <div>
          <label>Time Step: </label>
          <input
            type="number"
            value={timeStep}
            onChange={handleTimeStepChange}
          />
        </div>
    </div>
  );
};

export default Settings;