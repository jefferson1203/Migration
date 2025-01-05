import React from 'react';

const Controls = ({ isRunning, onStart, onStop }) => {
  return (
    <div>
      <button onClick={onStart} disabled={isRunning}>Start</button>
      <button onClick={onStop} disabled={!isRunning}>Stop</button>
    </div>
  );
};

export default Controls;