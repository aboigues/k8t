# Correctif proposé par Mirador

## Contexte
Le workflow « Security Scans » (job *Go Vulnerability Check*) échoue avec `exit code 3` : govulncheck détecte 5 vulnérabilités atteignables par le code.

## Vulnérabilités détectées
| ID | Paquet | Trouvé | Corrigé |
|----|--------|--------|---------|
| GO-2026-5856 | crypto/tls (stdlib) | go1.25.10 | go1.25.12 |
| GO-2026-5039 | net/textproto (stdlib) | go1.25.10 | go1.25.11 |
| GO-2026-5038 | mime (stdlib) | go1.25.10 | go1.25.11 |
| GO-2026-5037 | crypto/x509 (stdlib) | go1.25.10 | go1.25.11 |
| GO-2026-5026 | golang.org/x/net/idna | v0.53.0 | v0.55.0 |

Traces d'appel principales : `pkg/k8s/pods.go:245` et `:261` (`Client.GetContainerLogs`), `pkg/k8s/events.go:40` (`Client.GetPodEvents`), `pkg/output/audit.go:150` (`output.SimpleLog`).

## Correctif proposé
1. **Corriger les vulnérabilités de la bibliothèque standard** en forçant une toolchain Go >= 1.25.12. Dans `go.mod`, ajouter/mettre à jour la directive :
   ```
   go 1.25.12
   toolchain go1.25.12
   ```
   Le workflow utilise `actions/setup-go@v5` avec `go-version-file: go.mod` : la nouvelle version sera automatiquement récupérée. (Vérifier aussi que le cache setup-go recompilera bien avec cette version.)
2. **Mettre à jour golang.org/x/net** vers >= v0.55.0 :
   ```bash
   go get golang.org/x/net@v0.55.0
   go mod tidy
   ```
3. Relancer localement `govulncheck ./...` pour confirmer l'absence de vulnérabilités atteignables.

## Notes
- `go.mod`/`go.sum` n'ont pas été inclus intégralement dans cette PR car leur contenu complet n'est pas disponible dans le log ; les commandes ci-dessus produisent les modifications correctes de manière reproductible.
- Point d'hygiène complémentaire (hors périmètre bloquant) : migrer `github/codeql-action@v3` vers `v4` et les actions ciblant Node 20 vers des versions Node 24, comme signalé par les warnings.

## Détail du correctif

```
Mettre à jour la toolchain Go vers 1.25.12 dans go.mod (directives `go 1.25.12` et `toolchain go1.25.12`) et bumper golang.org/x/net vers v0.55.0 via `go get golang.org/x/net@v0.55.0 && go mod tidy`, puis revalider avec `govulncheck ./...`.
```

> Correctif à compléter/appliquer par un responsable (Mirador n'a pas pu le matérialiser automatiquement).
