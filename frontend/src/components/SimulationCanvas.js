import React, { useRef, useEffect } from 'react';

const SimulationCanvas = ({ simulationState }) => {
    const canvasRef = useRef(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        const width = canvas.width;
        const height = canvas.height;
        const worldSize = simulationState.worldSize;
        const scaleX = width / worldSize;
        const scaleY = height / worldSize;
        const birdSize = 5;
        const obstacleSize = 3;
        const resourceSize = 5;

        ctx.clearRect(0, 0, width, height); // Clear the canvas
        
        if (simulationState.obstacles && simulationState.obstacles.length > 0) {
            simulationState.obstacles.forEach(obstacle => {
                ctx.beginPath();
                ctx.arc(obstacle.Position[0] * scaleX, obstacle.Position[1] * scaleY, obstacle.Radius * scaleX * obstacleSize, 0, 2 * Math.PI);
                ctx.fillStyle = 'gray';
                ctx.fill();
                ctx.closePath();
            });
        }

        if (simulationState.resources && simulationState.resources.length > 0) {
            simulationState.resources.forEach(resource => {
                ctx.beginPath();
                ctx.arc(resource.Position[0] * scaleX, resource.Position[1] * scaleY, resourceSize, 0, 2 * Math.PI);
                if (resource.Type === "food") {
                    ctx.fillStyle = 'orange';
                } else {
                    ctx.fillStyle = 'lightblue';
                }
                ctx.fill();
                ctx.closePath();
            });
        }

        if (simulationState.birds && simulationState.birds.length > 0) {
            simulationState.birds.forEach(bird => {
                ctx.beginPath();
                ctx.arc(bird.Position[0] * scaleX, bird.Position[1] * scaleY, birdSize, 0, 2 * Math.PI);
                if (bird.State === "resting") {
                    ctx.fillStyle = 'blue';
                } else if (bird.State === "searchingFood") {
                    ctx.fillStyle = 'purple';
                } else {
                    ctx.fillStyle = 'green';
                }
                ctx.fill();
                ctx.closePath();
            });
        }

        console.log("simulation state:", simulationState); // debug

    }, [simulationState]);

    return <canvas ref={canvasRef} width="800" height="600" style={{ border: '1px solid black' }} />;
};

export default SimulationCanvas;