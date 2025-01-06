# Bird Migration Simulation

Ce projet est une simulation multi-agent de la migration d'oiseaux, développée avec Go (backend) et React (frontend). Il permet de visualiser le déplacement d'oiseaux dans un environnement simulé, d'ajuster les paramètres de la simulation, et de sauvegarder l'état de celle-ci.

## Fonctionnalités Principales

*   **Simulation de Migration :** Les oiseaux (agents) se déplacent dans un environnement simulé avec des règles simples (migration, repos, recherche de nourriture, évitement d'obstacles).
*   **Visualisation en Temps Réel :** Le frontend React affiche l'état de la simulation en temps réel.
*   **Contrôles :** Boutons pour démarrer et arrêter la simulation.
*   **Paramètres Ajustables :** Possibilité de modifier la vitesse de la simulation et la taille du monde.
*   **Persistance des Données :** Sauvegarde et chargement de l'état de la simulation via SQLite.
*   **Indicateur de Collisions :** Affiche le nombre de collisions entre oiseaux.
*   **Interface Utilisateur Dynamique :** L'interface a un thème inspiré de Barbie, avec des couleurs vives et des animations subtiles.

## Architecture Technique

*   **Backend (Go/Gin) :**
    *   Utilise le framework `Gin` pour créer une API REST.
    *   Implémente la logique de la simulation, la gestion des agents, et les interactions avec l'environnement.
    *   Stocke et récupère l'état de la simulation avec `SQLite`.
    *   Utilise des channels (au lieu de mutexes) pour la communication entre les goroutines.
*   **Frontend (React) :**
    *   Utilise `React` pour la construction des composants et la gestion de l'état.
    *   Utilise `axios` pour communiquer avec l'API REST du backend.
    *   Affiche le monde de la simulation avec un élément `canvas`.
    *   Utilise `useEffect` pour faire les appels à l'API de manière efficace.

## Installation et Lancement

### Prérequis

*   [Go](https://go.dev/dl/) (version 1.18 ou supérieure)
*   [Node.js](https://nodejs.org/en/download/) (version 16 ou supérieure) et npm (ou yarn)
*   [Git](https://git-scm.com/downloads) (si vous souhaitez cloner le repo)
*   Un éditeur de code (ex: VS Code, GoLand)
*   Une installation de SQLite pour pouvoir utiliser la base de données en backend

### Lancer l'application

Pour lancer l'application, suivez les étapes ci-dessous :

1. **Clonez le repository :**
 ```bash
 git clone https://github.com/ton_utilisateur/bird-migration-simulation.git
 ```

2. **Démarrez le Backend :**
    * Naviguez vers le dossier `backend` :

     ```bash
     cd bird-migration-simulation/backend
     ```
    *  **Créez un fichier `.env` :**
        *  Créez un fichier `.env` à la racine du dossier `backend`, et ajoutez les variables suivantes :

         ```env
         PORT=8080
         SIMULATION_SPEED=100
         WORLD_SIZE=1000
         INITIAL_BIRDS=50
         ENVIRONMENT_SIZE=100
         DB_PATH=simulation.db
         OBSTACLE_COUNT=5
         RESOURCE_COUNT=5
         ```

    *  **Exécutez le serveur backend :**
    ```bash
    go run main.go
     ```

    Le backend devrait démarrer et être accessible sur `http://localhost:8080`.

3.  **Démarrez le Frontend :**
    *   Naviguez vers le dossier `frontend` :

         ```bash
         cd bird-migration-simulation/frontend
         ```
    *   **Installez les dépendances :**

         ```bash
         npm install
         ```
    *   **Créez un fichier `.env` :**
        *   Créez un fichier `.env` à la racine du dossier `frontend/src` et ajoutez la ligne suivante :

         ```env
         REACT_APP_BACKEND_URL=http://localhost:8080
         ```
    * **Exécutez l'application React :**
    ```bash
    npm start
    ```

Le frontend devrait démarrer et être accessible dans votre navigateur à l'adresse `http://localhost:3000`.

### Utilisation

*   **Démarrer la Simulation :** Cliquez sur le bouton "Start" pour lancer la simulation.
*   **Arrêter la Simulation :** Cliquez sur le bouton "Stop" pour arrêter la simulation.
*   **Modifier les paramètres :** Utilisez le formulaire "Settings" pour modifier la vitesse de la simulation et la taille du monde.
*   **Visualiser la Simulation :** Observez le mouvement des oiseaux dans la zone de simulation.
*   **Charger une Sauvegarde :** Cliquez sur le bouton "Load" pour récupérer la dernière sauvegarde depuis la base de données.
*   **Enregistrer la Sauvegarde :** Cliquez sur le bouton "Save" pour enregistrer l'état actuel de la simulation dans la base de données.

## Structure du Code

```
bird-migration-simulation/
├── backend/           # Backend Go avec Gin
│   ├── main.go        # Fichier principal du backend
│   ├── go.mod         # Dépendances Go
│   ├── go.sum
│   └── config.go      # Configuration du backend
├── frontend/          # Frontend React
│   ├── src/           
│   │   ├── App.js       # Composant principal de l'application
│   │   ├── components/  # Composants React
│   │   │   ├── SimulationCanvas.js # Canvas pour la simulation
│   │   │   ├── Controls.js      # Contrôles de la simulation
│   │   │   ├── Settings.js      # Panneau de paramètres
│   │   ├── utils/      # Fonctions utilitaires
│   │   │   ├── api.js  # Fonctions pour l'API backend
│   │   ├── styles/     # Styles
│   │   │   ├── styles.css # Styles globaux
│   │   └── index.js      # Point d'entrée du frontend
│   ├── package.json   # Dépendances Node
│   ├── package-lock.json
│   └── .env           # Variables d'environnement
└── README.md        # Informations du projet
```

## Contributions

Les contributions sont les bienvenues ! Voici quelques idées d'amélioration :

*   Ajouter des comportements plus complexes pour les oiseaux (vol en groupe, évitement des obstacles, recherche de nourriture).
*   Améliorer la visualisation avec un environnement plus détaillé.
*   Ajouter des interfaces pour l'ajout d'obstacles et de ressources par l'utilisateur.
*   Ajouter des tests unitaires et d'intégration.
*   Mettre en place une sauvegarde et un chargement de l'état plus sophistiqués (gestion de différentes sauvegardes, etc).
*   Améliorer l'UI pour une expérience utilisateur plus agréable.

Pour contribuer, suivez les étapes suivantes :

1.  **Fork** le dépôt.
2.  **Clonez** votre fork.
3.  **Créez une branche** pour vos modifications.
4.  **Appliquez vos modifications**.
5.  **Poussez** votre branche sur votre fork.
6.  **Ouvrez une Pull Request** vers le dépôt original.

## Auteur

Jefferson MBOUOPDA
