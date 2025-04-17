# Forum

Eléves : Mathys Urban, Vittorio Gandossi, Cartagena Lucas, Bro Max

## Avant toute chose, on va installer mckert :
- Ouvrir VSCode
- Supprimer les fichiers "localhost.pem" et "localhost-key.pem" du projet VScode

- Ouvrir un nouveau terminal
- générer le nouveau certificat sur ton PC avec les commandes "mkcert -install" et "mkcert localhost"
- Ajouter le certificat à ton système avec ces 2 commandes (Linux uniquement) :
sudo cp "$(mkcert -CAROOT)/rootCA.pem" /usr/local/share/ca-certificates/mkcert-root.crt
sudo update-ca-certificates


## Pour allumer le serveur :
- Installer docker avec les commandes suivantes :
  sudo apt install apt-transport-https ca-certificates curl software-properties-common
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
  sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  sudo apt install docker-ce
- Taper: sudo docker compose up --build

## Ensuite pour accéder au site internet :
- Ouvrir une page internet (Firefox de préférence)
- Taper dans la barre d'URL, le lien suivant :
http://localhost:8080/

## Connexion pour administrateur :
 e-mail : test@outlook.fr
 mot de passe : Password

## Lien vers notre diaporama avec (entre autres) le diagramme de Gantt et le schéma fonctionnel :
https://prezi.com/view/iMdE0OV5AtywX1Ss725c/
