import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './Dashboard.css'; // Add a CSS file for styling

const Dashboard = ({ onUpdate }) => {
  const [factors, setFactors] = useState({
    temperature: 20.0,
    foodAvailability: 1.0,
    predatorPresence: 0.1, // Set initial value to 0.1
  });

  const [zones, setZones] = useState([]);
  const [selectedZone, setSelectedZone] = useState(null);

  useEffect(() => {
    const fetchFactors = async () => {
      try {
        const response = await axios.get('http://localhost:8080/environment');
        setFactors(response.data);
      } catch (error) {
        console.error('Failed to fetch environmental factors:', error);
      }
    };
    fetchFactors();

    const fetchZones = async () => {
      try {
        const response = await axios.get('http://localhost:8080/zones');
        console.log('Fetched zones:', response.data); // Log fetched zones
        setZones(response.data);
        setSelectedZone(response.data[0]?.id || null); // Set initial selected zone
      } catch (error) {
        console.error('Failed to fetch zones:', error);
      }
    };
    fetchZones();
  }, []);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFactors((prevFactors) => ({
      ...prevFactors,
      [name]: parseFloat(value),
    }));
  };

  const handleZoneChange = (id, name, value) => {
    setZones((prevZones) =>
      prevZones.map((zone) =>
        zone.id === id ? { ...zone, [name]: parseFloat(value) } : zone
      )
    );
  };

  const handleZoneSelect = (e) => {
    setSelectedZone(parseInt(e.target.value, 10));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await axios.post('http://localhost:8080/environment', factors);
      await axios.post('http://localhost:8080/zones', zones);
      onUpdate(factors);
    } catch (error) {
      console.error('Failed to update environmental factors:', error);
    }
  };

  return (
    <div className="dashboard-container">
      <div className="dashboard">
        <h2>Environmental Factors</h2>
        <form onSubmit={handleSubmit} className="dashboard-form">
          <div className="form-group">
            <label>Temperature:</label>
            <input
              type="number"
              name="temperature"
              value={factors.temperature}
              onChange={handleChange}
              className="form-control"
            />
          </div>
          <div className="form-group">
            <label>Food Availability:</label>
            <input
              type="number"
              name="foodAvailability"
              value={factors.foodAvailability}
              onChange={handleChange}
              className="form-control"
            />
          </div>
          <div className="form-group">
            <label>Predator Presence:</label>
            <input
              type="number"
              name="predatorPresence"
              value={factors.predatorPresence}
              onChange={handleChange}
              className="form-control"
            />
          </div>
          <h3>Zones</h3>
          <div className="form-group">
            <label>Select Zone:</label>
            <select value={selectedZone} onChange={handleZoneSelect} className="form-control">
              {zones.map((zone) => (
                <option key={zone.id} value={zone.id}>
                  Zone {zone.id}
                </option>
              ))}
            </select>
          </div>
          {selectedZone !== null && (
            <div className="form-group">
              <h4>Zone {selectedZone}</h4>
              <label>Temperature:</label>
              <input
                type="number"
                value={zones.find((zone) => zone.id === selectedZone)?.temperature || ''}
                onChange={(e) => handleZoneChange(selectedZone, 'temperature', e.target.value)}
                className="form-control"
              />
              <label>Food Availability:</label>
              <input
                type="number"
                value={zones.find((zone) => zone.id === selectedZone)?.foodAvailability || ''}
                onChange={(e) => handleZoneChange(selectedZone, 'foodAvailability', e.target.value)}
                className="form-control"
              />
              <label>Predator Presence:</label>
              <input
                type="number"
                value={zones.find((zone) => zone.id === selectedZone)?.predatorPresence || ''}
                onChange={(e) => handleZoneChange(selectedZone, 'predatorPresence', e.target.value)}
                className="form-control"
              />
            </div>
          )}
          <button type="submit" className="btn btn-primary">Update Factors</button>
        </form>
      </div>
      <div className="zone-summary">
        <h3>Zone Summary</h3>
        {zones.map((zone) => (
          <div key={zone.id} className="zone-summary-item">
            <h4>Zone {zone.id}</h4>
            <p>Temperature: {zone.temperature}°C</p>
            <p>Food Availability: {zone.foodAvailability}</p>
            <p>Predator Presence: {zone.predatorPresence}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Dashboard;
