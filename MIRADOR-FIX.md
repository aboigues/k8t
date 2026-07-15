# Correctif proposé par Mirador

## Contexte
Le workflow « Security Scans » échoue sur l'étape `govulncheck ./...` (exit code 3). Le scan détecte 5 vulnérabilités atteignables par le code (traces via `pkg/k8s/pods.go`, `pkg/k8s/events.go` et `pkg/output/audit.go`).

## Vulnérabilités détectées et versions correctrices
| ID | Paquet | Trouvé | Corrigé |
|----|--------|--------|---------|
| GO-2026-5856 | crypto/tls (stdlib) | go1.25.10 | go1.25.12 |
| GO-2026-5039 | net/textproto (stdlib) | go1.25.10 | go1.25.11 |
| GO-2026-5038 | mime (stdlib) | go1.25.10 | go1.25.11 |
| GO-2026-5037 | crypto/x509 (stdlib) | go1.25.10 | go1.25.11 |
| GO-2026-5026 | golang.org/x/net/idna | v0.53.0 | v0.55.0 |

## Correctif proposé
1. **Bibliothèque standard** : relever la version du toolchain Go à **1.25.12** (couvre les 4 correctifs stdlib). Dans `go.mod`, mettre à jour la directive `go` / `toolchain` :
   ```
   go 1.25.12
   ```
   (ou `toolchain go1.25.12` si vous utilisez une directive séparée). Le workflow s'appuie sur `actions/setup-go@v5` avec `go-version-file: go.mod`, donc la mise à jour de `go.mod` suffit à faire compiler/scanner avec la bonne version.
2. **Dépendance x/net** : mettre à jour vers v0.55.0 :
   ```
   go get golang.org/x/net@v0.55.0
   go mod tidy
   ```
3. Régénérer `go.sum` via `go mod tidy` puis vérifier localement :
   ```
   govulncheck ./...
   ```

## Note
Les fichiers `go.mod`/`go.sum` ne sont pas fournis intégralement dans cette PR automatique car leur contenu complet (chemin de module, ensemble des dépendances et sommes de contrôle) n'est pas connu depuis le log ; `go.sum` doit être régénéré par `go mod tidy` pour garantir des hachages corrects. Merci d'appliquer les commandes ci-dessus.

## Détail du correctif

```
Relever la directive Go à 1.25.12 dans go.mod, exécuter `go get golang.org/x/net@v0.55.0` puis `go mod tidy` pour régénérer go.sum, et valider avec `govulncheck ./...`.
```

> Correctif à compléter/appliquer par un responsable (Mirador n'a pas pu le matérialiser automatiquement).
