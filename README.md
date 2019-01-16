# TCChat (Go)

## Installation, lancement
Le paquet tui-go est nécessaire pour lancer le Client. On peut spécifier le nom et l'adresse, sinon ils auront des valeurs par défaut:
```
$ go get github.com/marcusolsson/tui-go
$ git clone git@github.com:gforien/tcchat.git
$ go run Server.go -address 'localhost:2000' -name 'Le serveur de TC'
$ go run Client.go -address 'localhost:2000' -name 'tom'
```

Si pour une raison quelconque la version TUI du Client ne marche pas:
- essayer de lancer le Client dans les tty2, tty3, etc.
- switcher sur la branche Commands qui fait des affichages classiques.


## Fonctionnalités ajoutées
Nous avons ajouté 3 fonctionnalités essentielles à un chat : envoyer/recevoir des messages privés, demander la liste des utilisateurs connectés, et bannir un utilisateur.
Les contraintes du protocole spécifiées dans le sujet ont été conservées pour ces nouvelles fonctionnalités (les noms d'utilisateurs ne doivent **pas contenir de \t**, les messages doivent être **inférieurs à 140 caractères**).

Le client envoie donc :
1) `TCCHAT_PRIVATE\t<nickname>\t<recipient>\t<message_payload>\n`
2) `TCCHAT_USERS\n`
3) `TCCHAT_BAN\t<nickname>\t<user_to_ban>\n`

Auquels le serveur répond respectivement :
1) `TCCHAT_PERSONAL\t<Nickname>\t<Payload>\n`
2) `TCCHAT_USERLIST\t<user1>\r...\r<userN>\n`.
3) `TCCHAT_USERBAN\t<user_who_banned>\t<user_banned>\n`


## Annexe - Rappel du cycle de vie d'un projet git
##### 1) Simple commit
```
$ git status
-> Sur la branche master
$ vim fichier1 fichier2                                 # modifications
$ git add fichier1 fichier2
$ git commit -m "Commit sur la branche principale"
$ git pull && git push
```
##### 2) Création d'une nouvelle branche, commits, et fermeture de la branche
```
$ git checkout -b gui
-> Sur la nouvelle branche gui

$ vim fichier1 fichier2                                 # modifications
$ git add fichier1 fichier2
$ git commit -m "Commit sur la branche gui"
$ git push -u                                           # -u si la branche n'existe pas déjà sur le serveur

$ vim fichier1 fichier2                                 #  nouvelles modifications
$ git add fichier1 fichier2
$ git commit -m "2e commit sur la branche gui"
$ git push                                              # cette fois la branche existe déjà sur le serveur

$ git checkout master
-> Sur la branche master
$ git merge --no-ff gui                                 # on rapatrie la branche master sur la branche gui
$ git branch -d gui                                     # supprime la branche gui
$ git pull && git push
```
##### 3) Sauvegarder le travail en cours pour travailler sur la branche principale
```
$ git st
-> Changes not staged for commit : fichier1
$ git stash                                             # enregistre fichier1 en mémoire et restaure le dernier commit

$ vim fichier1 fichier2                                 # modifications
$ git add fichier1 fichier2
$ git commit -m "Commit sur la branche principale"
$ git pull && git push

$ git stash pop                                           # restaure fichier1
```

##
###### Rémi ARLANDA, Gabriel FORIEN, Rémi POLIDO, Romain THAURONT <br/>3TC INSA Lyon
![Logo INSA Lyon](https://upload.wikimedia.org/wikipedia/commons/b/b9/Logo_INSA_Lyon_%282014%29.svg)
