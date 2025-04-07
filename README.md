# Forum

## Avant toute chose, on va installer mckert  :
- Ouvrir VSCode  
- Supprimer les fichiers "localhost.pem" et "localhost-key.pem" du projet VScode

- Ouvrir un nouveau terminal  
- génèrer le nouveau certificat sur ton PC avec les commandes "mkcert -install" et "mkcert localhost"
- Ajouter le certificat à ton système avec ces 2 commandes (Linux uniquement) :
        sudo cp "$(mkcert -CAROOT)/rootCA.pem" /usr/local/share/ca-certificates/mkcert-root.crt
        sudo update-ca-certificates


## Pour allumer le serveur :
- Taper: go run server/main.go  

## Ensuite pour accéder au site internet :
- Ouvrir une page internet (Firefox de préférence)  
- Taper dans la barre d'URL, le lien suivant :  
  http://localhost:8080
