import React, { useRef, useEffect } from "react";

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
        const obstacleSize = 3;
        const resourceSize = 5;

        // Clear canvas before redrawing
        ctx.clearRect(0, 0, width, height);
        console.log("Updated simulationState:", simulationState);

        // Draw obstacles
        simulationState.obstacles?.forEach((obstacle) => {
            ctx.beginPath();
            ctx.arc(
                obstacle.position[0] * scaleX,
                obstacle.position[1] * scaleY,
                obstacle.radius * scaleX * obstacleSize,
                0,
                2 * Math.PI
            );
            ctx.fillStyle = "gray";
            ctx.fill();
            ctx.closePath();
        });

        // Draw resources
        simulationState.resources?.forEach((resource) => {
            ctx.beginPath();
            ctx.arc(
                resource.position[0] * scaleX,
                resource.position[1] * scaleY,
                resourceSize,
                0,
                2 * Math.PI
            );
            ctx.fillStyle = resource.type === "food" ? "orange" : "lightblue";
            ctx.fill();
            ctx.closePath();
        });

        // Draw birds
        simulationState.birds?.forEach((bird) => {
            ctx.beginPath();
            ctx.arc(
                bird.position[0] * scaleX,
                bird.position[1] * scaleY,
                birdSize,
                0,
                2 * Math.PI
            );
            ctx.fillStyle =
                bird.state === "resting"
                    ? "blue"
                    : bird.state === "searchingFood"
                    ? "purple"
                    : "green";
            ctx.fill();
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