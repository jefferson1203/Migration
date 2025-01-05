import axios from 'axios';
  
const API_URL = process.env.REACT_APP_BACKEND_URL || 'http://localhost:8080';

export const fetchSimulationState = async () => {
  try {
      const response = await axios.get(`${API_URL}/simulation`);
      return response.data;
  } catch (error) {
      console.error("Error fetching simulation state:", error);
      throw error;
  }
};

export const startSimulation = async () => {
  try {
      await axios.post(`${API_URL}/simulation/start`);
  } catch (error) {
      console.error("Error starting simulation:", error);
      throw error;
  }
};

export const stopSimulation = async () => {
  try {
      await axios.post(`${API_URL}/simulation/stop`);
  } catch (error) {
      console.error("Error stopping simulation:", error);
      throw error;
  }
};

export const fetchSimulationConfig = async () => {
    try {
        const response = await axios.get(`${API_URL}/simulation/config`);
        return response.data;
    } catch (error) {
        console.error("Error fetching simulation config:", error);
        throw error;
    }
};

export const updateSimulationConfig = async (config) => {
    try {
        await axios.post(`${API_URL}/simulation/config`, config);
    } catch (error) {
        console.error("Error updating simulation config:", error);
        throw error;
    }
};

export const setTimeStep = async (timeStep) => {
    try {
        await axios.post(`${API_URL}/simulation/time-step`, { timeStep });
    } catch (error) {
        console.error("Error setting time step:", error);
        throw error;
    }
};

export const fetchTimeStep = async () => {
    try {
        const response = await axios.get(`${API_URL}/simulation/time-step`);
        return response.data;
    } catch (error) {
        console.error("Error fetching time step:", error);
        throw error;
    }
};

export const saveSimulation = async () => {
    try {
        await axios.post(`${API_URL}/simulation/save`);
    } catch (error) {
        console.error("Error saving simulation:", error);
        throw error;
    }
};

export const loadSimulation = async () => {
    try {
        const response = await axios.get(`${API_URL}/simulation/load`);
        return response.data;
    } catch (error) {
        console.error("Error loading simulation:", error);
        throw error;
    }
};