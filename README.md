# TCChat
Projet de chat client-serveur en Go.
3TC INSA Lyon

## Cycle de vie classique d'un projet git
- Voir mes modifications
- ```git status```
- Créer un _commit_ c'est-à-dire un ensemble de modifications cohérentes entre elles. Par exemple une nouvelle fonctionnalité ajoutée; ou un bug corrigé. Un commit contient **obligatoirement** un message.
- ```git add fichier1 fichier2```
- ```git commit -m "Bug sur le slice corrigé"```
- Downloader les modifications que n'importe qui aurait pu ajouter au projet, et enfin uploader ses propres modifications
- ```git pull```
- ```git push```
- Créer une branche, càd une "copie" du projet sur laquelle on devellope sans changer la branche principale (souvent pour conserver une version qui marche et avoir une version de développement)
- ```git checkout -b [nom de la nouvelle branche]```
- Changer de branche /!\ toujours |pull - commit - push| avant de changer de branches
- ```git checkout [nom branche]```
