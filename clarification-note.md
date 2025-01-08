# Note de Clarification

## Objectif du Projet

Le projet Bird Migration Simulation a pour objectif de simuler le comportement migratoire des oiseaux dans un environnement virtuel. Cette simulation permet d'observer et d'analyser les interactions entre les oiseaux et leur environnement, ainsi que les effets de différents facteurs environnementaux sur leur comportement.

## Composants Principaux

### Backend (Go)

Le backend est responsable de la logique de la simulation. Il gère les agents (oiseaux), les obstacles, les ressources, et les interactions entre ces éléments. Le backend utilise le framework Gin pour exposer une API REST permettant de contrôler la simulation et de récupérer son état.

### Frontend (React)

Le frontend est responsable de l'affichage de la simulation en temps réel et de l'interaction avec l'utilisateur. Il utilise React pour construire l'interface utilisateur et axios pour communiquer avec l'API REST du backend. Le frontend permet de démarrer et d'arrêter la simulation, de modifier les paramètres, et de visualiser l'état actuel de la simulation.

## Fonctionnalités Clés

- **Simulation de Migration :** Les oiseaux se déplacent selon des règles de migration, de repos, de recherche de nourriture, et d'évitement d'obstacles.
- **Visualisation en Temps Réel :** L'état de la simulation est affiché en temps réel sur le frontend.
- **Contrôles Utilisateur :** L'utilisateur peut démarrer et arrêter la simulation, ainsi que modifier les paramètres de la simulation.
- **Facteurs Environnementaux :** L'utilisateur peut ajuster la température, la disponibilité de la nourriture, et la présence de prédateurs pour influencer le comportement des oiseaux.

## Diagramme UML

Pour une vue d'ensemble des classes et de leurs relations, veuillez consulter le fichier [uml-diagram.md](./uml-diagram.md).

## Instructions d'Installation

Pour des instructions détaillées sur l'installation et le lancement du projet, veuillez consulter le fichier [Readme.md](./Readme.md).

## Contributions

Les contributions sont les bienvenues ! Pour plus de détails sur la manière de contribuer, veuillez consulter la section "Contributions" du fichier [Readme.md](./Readme.md).

## Auteur

Jefferson MBOUOPDA
