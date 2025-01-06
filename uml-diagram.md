# Diagramme UML du Projet Bird Migration Simulation

Ce document présente le diagramme UML du projet Bird Migration Simulation, illustrant les principales classes et leurs relations.

## Diagramme de Classes

```plaintext
+----------------+        +----------------+        +----------------+
|   Simulation   |        |      Bird      |        |    Obstacle    |
+----------------+        +----------------+        +----------------+
| - worldSize    |        | - id           |        | - id           |
| - birds        |        | - position     |        | - position     |
| - obstacles    |        | - state        |        | - size         |
| - resources    |        | - energy       |        +----------------+
| - collisionCount|       +----------------+
+----------------+        | + move()       |
| + start()      |        | + rest()       |
| + stop()       |        | + searchFood() |
| + update()     |        | + avoidObstacle()|
| + saveState()  |        +----------------+
| + loadState()  |
+----------------+

+----------------+        +----------------+        +----------------+
|    Resource    |        |      App       |        | SimulationCanvas|
+----------------+        +----------------+        +----------------+
| - id           |        | - simulationState|      | - simulationState|
| - position     |        +----------------+        +----------------+
| - type         |        | + render()     |        | + render()      |
+----------------+        +----------------+        +----------------+

+----------------+        +----------------+        +----------------+
|    Controls    |        |    Settings    |        |   Dashboard    |
+----------------+        +----------------+        +----------------+
| - isRunning    |        | - onUpdate     |        | - onUpdate     |
| - onStart      |        +----------------+        +----------------+
| - onStop       |        | + render()     |        | + render()     |
+----------------+        +----------------+        +----------------+
| + render()     |
+----------------+
```

### Description des Classes

#### Backend (Go)

- **Simulation**
  - Attributs : `worldSize`, `birds`, `obstacles`, `resources`, `collisionCount`
  - Méthodes : `start()`, `stop()`, `update()`, `saveState()`, `loadState()`

- **Bird**
  - Attributs : `id`, `position`, `state`, `energy`
  - Méthodes : `move()`, `rest()`, `searchFood()`, `avoidObstacle()`

- **Obstacle**
  - Attributs : `id`, `position`, `size`

- **Resource**
  - Attributs : `id`, `position`, `type`

#### Frontend (React)

- **App**
  - Composants : `SimulationCanvas`, `Controls`, `Settings`, `Dashboard`
  - État : `simulationState`

- **SimulationCanvas**
  - Props : `simulationState`

- **Controls**
  - Props : `isRunning`, `onStart`, `onStop`

- **Settings**
  - Props : `onUpdate`

- **Dashboard**
  - Props : `onUpdate`

### Relations

- **Simulation** contient plusieurs **Bird**, **Obstacle**, et **Resource**.
- **App** gère l'état global de la simulation et passe les données aux composants enfants.
- **SimulationCanvas** affiche l'état de la simulation.
- **Controls** permet de démarrer et d'arrêter la simulation.
- **Settings** permet de modifier les paramètres de la simulation.
- **Dashboard** permet de modifier les facteurs environnementaux.
