# Tests Manuels k8t

Ce répertoire contient les manifests et scripts pour tester k8t avec Minikube.

## Structure

```
tests/manual/
├── README.md              # Ce fichier
├── TEST_PLAN.md          # Plan de test détaillé
├── manifests/            # Manifests Kubernetes pour les tests
│   ├── 01-test-image-not-found.yaml
│   ├── 02-test-auth-failure.yaml
│   ├── 03-test-network-issue.yaml
│   ├── 04-test-manifest-error.yaml
│   ├── 05-test-success.yaml
│   └── 06-test-multiple-containers.yaml
└── scripts/              # Scripts utilitaires (à créer)
```

## Quick Start

### 1. Prérequis
```bash
# Vérifier que minikube est installé
minikube version

# Vérifier que kubectl est installé
kubectl version --client

# Compiler k8t
cd /mnt/c/Users/PC/OneDrive/Documents/git/k8t
make build
```

### 2. Démarrer Minikube
```bash
minikube start
```

### 3. Déployer les pods de test
```bash
kubectl apply -f tests/manual/manifests/
```

### 4. Attendre les erreurs ImagePullBackOff
```bash
# Surveiller les pods (Ctrl+C pour arrêter)
kubectl get pods -w

# Ou vérifier le statut
kubectl get pods
```

### 5. Tester k8t
```bash
# Test basique
./bin/k8t analyze imagepullbackoff test-image-not-found

# Avec JSON
./bin/k8t analyze imagepullbackoff test-image-not-found --output json

# Avec YAML
./bin/k8t analyze imagepullbackoff test-auth-failure --output yaml

# Mode verbose
./bin/k8t analyze imagepullbackoff test-network-issue --verbose
```

## Scénarios de Test

| Pod | Scénario | Root Cause Attendu | Severity |
|-----|----------|-------------------|----------|
| `test-image-not-found` | Image inexistante | IMAGE_NOT_FOUND | HIGH |
| `test-auth-failure` | Image privée sans creds | AUTHENTICATION_FAILURE | HIGH |
| `test-network-issue` | Registry inaccessible | NETWORK_ISSUE | MEDIUM |
| `test-manifest-error` | Architecture incompatible | MANIFEST_ERROR | MEDIUM |
| `test-success` | Image valide | Aucun problème | N/A |
| `test-multiple-containers` | Plusieurs containers | IMAGE_NOT_FOUND | HIGH |

## Commandes Utiles

### Voir les événements d'un pod
```bash
kubectl describe pod test-image-not-found
kubectl get events --field-selector involvedObject.name=test-image-not-found
```

### Forcer la recréation d'un pod
```bash
kubectl delete pod test-image-not-found
kubectl apply -f tests/manual/manifests/01-test-image-not-found.yaml
```

### Nettoyer tous les pods de test
```bash
kubectl delete -f tests/manual/manifests/
```

### Vérifier les logs k8t (avec verbose)
```bash
./bin/k8t analyze imagepullbackoff test-image-not-found --verbose 2>&1 | tee test.log
```

## Validation des Résultats

### 1. Vérifier le root cause
```bash
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq -r '.findings[0].root_cause'
```

### 2. Vérifier la severity
```bash
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq -r '.findings[0].severity'
```

### 3. Compter les steps de remediation
```bash
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq '.findings[0].remediation_steps | length'
```

### 4. Vérifier le parsing de l'image
```bash
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq '.findings[0].image_references[0]'
```

## Troubleshooting

### Minikube ne démarre pas
```bash
# Supprimer et recréer le cluster
minikube delete
minikube start
```

### Les pods ne passent pas en ImagePullBackOff
```bash
# Attendre plus longtemps (peut prendre 1-2 minutes)
kubectl get pods -w

# Forcer le pull
kubectl delete pod test-image-not-found
kubectl apply -f tests/manual/manifests/01-test-image-not-found.yaml
```

### k8t ne trouve pas le pod
```bash
# Vérifier que le pod existe
kubectl get pods

# Vérifier le namespace
kubectl get pods -n default

# Utiliser le flag namespace explicitement
./bin/k8t analyze imagepullbackoff test-image-not-found --namespace default
```

### Erreur de permission
```bash
# Vérifier les permissions RBAC
kubectl auth can-i get pods
kubectl auth can-i list events

# k8t devrait fonctionner avec les permissions par défaut de minikube
```

## Next Steps

Après avoir testé avec minikube:
1. Documenter les résultats dans TEST_PLAN.md
2. Prendre des screenshots des outputs
3. Noter les bugs ou améliorations nécessaires
4. Tester avec un vrai cluster Kubernetes si disponible
