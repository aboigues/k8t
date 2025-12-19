# Plan de Test Manuel - k8t ImagePullBackOff Analyzer

## Objectif
Tester l'analyseur k8t avec minikube en utilisant différents scénarios d'erreurs ImagePullBackOff.

## Prérequis
- Minikube installé et fonctionnel
- kubectl configuré
- k8t binary compilé (`make build`)

## Scénarios de Test

### Test 1: IMAGE_NOT_FOUND (404)
**Description**: Image qui n'existe pas dans Docker Hub
**Pod**: `test-image-not-found`
**Image**: `nginx:nonexistent-tag-12345`
**Résultat attendu**:
- Root Cause: IMAGE_NOT_FOUND
- Severity: HIGH
- Message d'erreur contenant "not found" ou "manifest unknown"

### Test 2: AUTHENTICATION_FAILURE (401/403)
**Description**: Image privée sans credentials
**Pod**: `test-auth-failure`
**Image**: `aboigues/private-test-image:latest` (image privée)
**Résultat attendu**:
- Root Cause: AUTHENTICATION_FAILURE
- Severity: HIGH
- Message d'erreur contenant "unauthorized" ou "authentication required"

### Test 3: NETWORK_ISSUE (Timeout/DNS)
**Description**: Registry inexistant ou inaccessible
**Pod**: `test-network-issue`
**Image**: `nonexistent-registry.example.com/nginx:latest`
**Résultat attendu**:
- Root Cause: NETWORK_ISSUE
- Severity: MEDIUM
- Message d'erreur contenant "timeout" ou "connection refused"

### Test 4: MANIFEST_ERROR (Architecture)
**Description**: Image non compatible avec l'architecture
**Pod**: `test-manifest-error`
**Image**: `arm32v7/nginx:latest` (si cluster est amd64)
**Résultat attendu**:
- Root Cause: MANIFEST_ERROR
- Severity: MEDIUM
- Message d'erreur contenant "no matching manifest" ou "unsupported platform"

### Test 5: SUCCESS (Image valide)
**Description**: Pod avec une image valide pour tester le cas "no issues"
**Pod**: `test-success`
**Image**: `nginx:latest`
**Résultat attendu**:
- Aucune erreur ImagePullBackOff
- Message: "No ImagePullBackOff issues found"

## Procédure de Test

### 1. Démarrage de Minikube
```bash
# Démarrer minikube
minikube start

# Vérifier le statut
minikube status
kubectl get nodes
```

### 2. Déploiement des Pods de Test
```bash
# Déployer tous les pods de test
kubectl apply -f tests/manual/manifests/

# Attendre que les pods soient en erreur (environ 30 secondes)
kubectl get pods -w
```

### 3. Vérification des Status
```bash
# Vérifier que les pods sont bien en ImagePullBackOff
kubectl get pods
kubectl describe pod test-image-not-found
kubectl describe pod test-auth-failure
kubectl describe pod test-network-issue
```

### 4. Tests avec k8t

#### Test 4.1: Output Text (par défaut)
```bash
./bin/k8t analyze imagepullbackoff test-image-not-found
```

**Vérifications**:
- ✓ Header avec nom du pod et namespace
- ✓ Root cause correctement identifiée
- ✓ Severity avec couleur appropriée (rouge pour HIGH)
- ✓ Liste des containers affectés
- ✓ Détails de l'image (registry, repository, tag)
- ✓ Steps de remediation numérotés et actionnables
- ✓ Événements récents affichés

#### Test 4.2: Output JSON
```bash
./bin/k8t analyze imagepullbackoff test-image-not-found --output json
```

**Vérifications**:
- ✓ JSON valide (peut être parsé avec jq)
- ✓ Structure complète avec tous les champs
- ✓ Pas de champs sensibles exposés (secrets redacted)

```bash
# Valider le JSON
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq .
```

#### Test 4.3: Output YAML
```bash
./bin/k8t analyze imagepullbackoff test-auth-failure --output yaml
```

**Vérifications**:
- ✓ YAML valide
- ✓ Lisible et bien formaté
- ✓ Même structure que JSON

#### Test 4.4: Flags et Options
```bash
# Test avec namespace explicite
./bin/k8t analyze imagepullbackoff test-network-issue --namespace default

# Test avec timeout personnalisé
./bin/k8t analyze imagepullbackoff test-manifest-error --timeout 60s

# Test avec verbose
./bin/k8t analyze imagepullbackoff test-image-not-found --verbose

# Test avec no-color
./bin/k8t analyze imagepullbackoff test-image-not-found --no-color
```

#### Test 4.5: Gestion d'Erreurs

**Pod inexistant**:
```bash
./bin/k8t analyze imagepullbackoff nonexistent-pod
```
**Résultat attendu**:
- Exit code: 3
- Message d'erreur clair
- Suggestions (kubectl get pods)

**Namespace inexistant**:
```bash
./bin/k8t analyze imagepullbackoff test-image-not-found --namespace nonexistent
```
**Résultat attendu**:
- Exit code: 2 ou 3
- Message d'erreur approprié

**Pod sans ImagePullBackOff**:
```bash
./bin/k8t analyze imagepullbackoff test-success
```
**Résultat attendu**:
- Exit code: 0
- Message: "No ImagePullBackOff issues found"

### 5. Tests de Précision

Pour chaque scénario, vérifier:

#### Root Cause Detection
```bash
# Vérifier que le root cause est correct
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq '.findings[0].root_cause'
# Attendu: "IMAGE_NOT_FOUND"

./bin/k8t analyze imagepullbackoff test-auth-failure -o json | jq '.findings[0].root_cause'
# Attendu: "AUTHENTICATION_FAILURE"

./bin/k8t analyze imagepullbackoff test-network-issue -o json | jq '.findings[0].root_cause'
# Attendu: "NETWORK_ISSUE"
```

#### Severity Mapping
```bash
# Vérifier que la severity est correcte
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq '.findings[0].severity'
# Attendu: "HIGH"
```

#### Image Reference Parsing
```bash
# Vérifier le parsing des références d'images
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq '.findings[0].image_references[0]'
```

**Vérifications**:
- ✓ Registry correctement extrait
- ✓ Repository correctement extrait
- ✓ Tag correctement extrait
- ✓ Container name correct

#### Remediation Steps
```bash
# Vérifier les steps de remediation
./bin/k8t analyze imagepullbackoff test-image-not-found -o json | jq '.findings[0].remediation_steps'
```

**Vérifications**:
- ✓ Au moins 3 steps fournis
- ✓ Steps actionnables et spécifiques
- ✓ Commandes concrètes incluses
- ✓ Pas de steps génériques inutiles

### 6. Tests de Performance

```bash
# Mesurer le temps d'exécution
time ./bin/k8t analyze imagepullbackoff test-image-not-found

# Attendu: < 5 secondes pour un pod simple
```

### 7. Tests de Sécurité

#### Secret Redaction
```bash
# Si un pod a des secrets dans les events, vérifier la redaction
./bin/k8t analyze imagepullbackoff test-auth-failure -o json | grep -i password
# Attendu: "[REDACTED]" ou aucun résultat
```

### 8. Nettoyage

```bash
# Supprimer les pods de test
kubectl delete -f tests/manual/manifests/

# Ou individuellement
kubectl delete pod test-image-not-found
kubectl delete pod test-auth-failure
kubectl delete pod test-network-issue
kubectl delete pod test-manifest-error
kubectl delete pod test-success

# Arrêter minikube (optionnel)
minikube stop
```

## Critères de Succès

### Fonctionnalité (MVP)
- [ ] Détection correcte des 8 root causes
- [ ] Severity mapping correct
- [ ] Remediation steps pertinents
- [ ] Parsing complet des image references
- [ ] Support des 3 formats de sortie (text, json, yaml)

### Qualité
- [ ] Pas de crash ou panic
- [ ] Gestion d'erreurs appropriée
- [ ] Exit codes corrects
- [ ] Messages d'erreur clairs et utiles
- [ ] Performance acceptable (< 10s par pod)

### Sécurité
- [ ] Secrets correctement redacted
- [ ] Pas d'information sensible dans les logs
- [ ] RBAC minimal requis (get pods, list events)

### UX
- [ ] Output text lisible et bien formaté
- [ ] Couleurs appropriées (rouge/jaune/vert)
- [ ] Help text clair
- [ ] Verbose mode informatif

## Bugs et Améliorations

Utiliser ce tableau pour tracker les bugs trouvés:

| # | Scénario | Problème | Priorité | Status |
|---|----------|----------|----------|--------|
| 1 | | | | |

## Notes

- Prendre des screenshots des différents outputs
- Noter les temps d'exécution
- Documenter tout comportement inattendu
- Tester sur différentes versions de Kubernetes si possible
