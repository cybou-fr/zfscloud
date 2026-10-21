# Étude de cas — ZFSCloud

**Projet personnel open source — Go, gRPC et architecture client–serveur de sauvegarde.**  
**État :** sources et guides de tests disponibles ; les mentions « Production Ready » dans le README ne constituent pas une certification ou un audit indépendant.

## 1. Problème et objectif

Concevoir une solution de sauvegarde dont le client pilote les opérations tandis qu’un serveur Linux gère des datasets ZFS isolés, les snapshots et la restauration des données.

## 2. Architecture retenue

- **Go :** services serveur, bibliothèques partagées, client CLI et interface desktop Gio.
- **gRPC / Protocol Buffers :** contrats explicites d’accès aux opérations et services administratifs.
- **ZFS sur Linux :** datasets utilisateurs, snapshots et chiffrement natif.
- **Transport et identités :** configuration TLS/mTLS et mécanismes d’authentification.
- **Clients :** surveillance des fichiers et parcours de sauvegarde/restauration.
- **Déploiement :** scripts d’installation et environnements Docker pour les tests.

**Flux :** client GUI/CLI → gRPC → services Backup/Admin → gestionnaire ZFS → datasets et snapshots.

[Protocole gRPC](../proto/backup.proto) · [Service Go](../server/grpc/backup_service.go) · [Composants ZFS](../server/zfs/)

## 3. Réalisation observable

Le code couvre le client, le serveur, des définitions de protocole, l’interface utilisateur et des scripts de tests. La documentation suit notamment des corrections dans le transport des passphrases et des travaux sur leur stockage local.

[État du chiffrement](../SECURITY_FIX_PROGRESS.md) · [Guide de tests](../LINUX_TESTING_GUIDE.md)

## 4. Choix et compromis

| Choix | Bénéfice | Limite |
| --- | --- | --- |
| Go côté client et serveur | Partage des modèles techniques | Spécificités OS à tester |
| gRPC | Contrats réseau typés | Compatibilité et gestion des erreurs |
| ZFS | Snapshots et isolation | Dépendance à Linux/ZFS côté serveur |
| Chiffrement et authentification | Protection des données | Audit du cycle de vie des clés |
| GUI et CLI | Plusieurs profils d’usage | Cohérence entre interfaces |

## 5. Validation et limites

La documentation indique des limites selon les plateformes concernant le stockage des clés et certaines étapes de sécurité. Elle ne prouve ni l’absence de vulnérabilités ni une utilisation en production chez un client. Aucun nouveau test n’a été réalisé pour cette étude.

## 6. Compétences illustrées

Go, architecture cloud et client–serveur, API/RPC, stockage ZFS, interfaces desktop, analyse de risques et documentation technique. Ce projet n’est pas présenté comme un système P2P décentralisé.

[Retour au dépôt](../README.md)
