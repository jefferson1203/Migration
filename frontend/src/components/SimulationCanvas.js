import React, { useRef, useEffect } from "react";
import birdImage from '../assets/bird.png'; // Add bird image
import mountainImage from '../assets/mountain.png'; // Add mountain image
import foodImage from '../assets/food.png'; // Add food image

const SimulationCanvas = ({ simulationState }) => {
    const canvasRef = useRef(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;

        const ctx = canvas.getContext("2d");
        const width = canvas.width;
        const height = canvas.height;
        const worldSize = simulationState.worldSize;
        const scaleX = width / worldSize;
        const scaleY = height / worldSize;
        const birdSize = 5;
        const obstacleSize = 5; // Increase obstacle size
        const resourceSize = 5;

        const birdImg = new Image();
        birdImg.src = birdImage;

        const mountainImg = new Image();
        mountainImg.src = mountainImage;

        const foodImg = new Image();
        foodImg.src = foodImage;

        // Clear canvas before redrawing
        ctx.clearRect(0, 0, width, height);
        console.log("Updated simulationState:", simulationState);

        // Draw obstacles as mountains
        simulationState.obstacles?.forEach((obstacle) => {
            ctx.drawImage(mountainImg, obstacle.position[0] * scaleX, obstacle.position[1] * scaleY, obstacle.radius * scaleX * obstacleSize, obstacle.radius * scaleY * obstacleSize);
        });

        // Draw resources
        simulationState.resources?.forEach((resource) => {
            if (resource.type === "food") {
                ctx.drawImage(foodImg, resource.position[0] * scaleX, resource.position[1] * scaleY, resourceSize * 4, resourceSize * 4);
                ctx.beginPath();
                ctx.arc(resource.position[0] * scaleX + resourceSize * 2, resource.position[1] * scaleY + resourceSize * 2, resourceSize * 2, 0, 2 * Math.PI);
                ctx.strokeStyle = "yellow";
                ctx.lineWidth = 2;
                ctx.stroke();
                ctx.closePath();
            } else {
                ctx.beginPath();
                ctx.arc(
                    resource.position[0] * scaleX,
                    resource.position[1] * scaleY,
                    resourceSize,
                    0,
                    2 * Math.PI
                );
                ctx.fillStyle = "lightblue";
                ctx.fill();
                ctx.closePath();
            }
        });

        // Draw birds
        simulationState.birds?.forEach((bird) => {
            ctx.drawImage(birdImg, bird.position[0] * scaleX, bird.position[1] * scaleY, birdSize * 4, birdSize * 4);
            ctx.beginPath();
            ctx.arc(bird.position[0] * scaleX + birdSize * 2, bird.position[1] * scaleY + birdSize * 2, birdSize * 2, 0, 2 * Math.PI);
            ctx.strokeStyle = bird.state === "migrating" ? "green" : bird.state === "resting" ? "blue" : "violet";
            ctx.lineWidth = 2;
            ctx.stroke();
            ctx.closePath();
        });
    }, [simulationState]);

    console.log("SimulationCanvas rendered", simulationState.birds[0]?.position);

    return (
        <div>
            <canvas
                ref={canvasRef}
                width={simulationState.worldSize}
                height={simulationState.worldSize}
                style={{ border: "1px solid black" }}
            />
        </div>
    );
};

export default React.memo(SimulationCanvas);