# Correctif proposé par Mirador

## Problème
Le workflow `Mirador Auto-fix` (run 29431597361) échoue à l'étape `gh pr create` avec :

```
pull request create failed: GraphQL: GitHub Actions is not permitted to create or approve pull requests (createPullRequest)
```

Toutes les étapes précédentes réussissent : la branche `mirador/autofix-test` est bien créée, le `go.mod`/`go.sum` mis à jour, commitée et poussée sur `origin`. Seule la création de la PR est bloquée.

## Cause
Par défaut le `GITHUB_TOKEN` d'Actions ne peut pas ouvrir de pull request lorsque le paramètre d'organisation/dépôt « Allow GitHub Actions to create and approve pull requests » est désactivé. Ce n'est pas un problème transitoire : c'est une restriction de permissions.

## Correctifs possibles (choisir l'un des deux)

### Option A — Activer le paramètre (aucun changement de code)
Dépôt (ou org) → **Settings → Actions → General → Workflow permissions** → cocher **« Allow GitHub Actions to create and approve pull requests »**. C'est la solution la plus simple si la politique de sécurité l'autorise.

### Option B — Utiliser un token dédié dans le workflow (recommandé si l'option A est interdite par politique)
Dans `.github/workflows/mirador-autofix.yml`, remplacer le `GITHUB_TOKEN` par un token disposant du droit de créer des PR, soit :
- un **GitHub App token** généré via `actions/create-github-app-token` (recommandé, permissions fines, pas d'expiration manuelle), soit
- un **PAT (fine-grained)** stocké dans un secret, ex. `secrets.MIRADOR_PAT`, avec la permission *Pull requests: write* et *Contents: write*.

Exemple pour l'étape de création de PR :

```yaml
      - name: Créer la Pull Request
        env:
          GH_TOKEN: ${{ secrets.MIRADOR_PAT }}   # PAT fine-grained (PR: write, Contents: write)
          BASE_BRANCH: main
          HEAD_BRANCH: ${{ env.HEAD_BRANCH }}
          TITRE: ${{ env.TITRE }}
          CORPS: ${{ env.CORPS }}
        run: |
          set -euo pipefail
          gh pr create --base "$BASE_BRANCH" --head "$HEAD_BRANCH" \
            --title "$TITRE" --body "${CORPS:-Correctif automatique matérialisé par Mirador.}"
```

Si l'on utilise `actions/create-github-app-token`, ajouter en amont :

```yaml
      - uses: actions/create-github-app-token@v1
        id: app-token
        with:
          app-id: ${{ vars.MIRADOR_APP_ID }}
          private-key: ${{ secrets.MIRADOR_APP_PRIVATE_KEY }}
      # puis GH_TOKEN: ${{ steps.app-token.outputs.token }}
```

> Note : la branche a déjà été poussée avec succès (`mirador/autofix-test`). Une fois le token/paramètre corrigé, la PR pourra être ouverte, soit par relance du workflow, soit manuellement via le lien affiché dans les logs.

## Nettoyage
Penser à supprimer la branche de test `mirador/autofix-test` (et éventuellement `mirador/fix-29382539931`) si elles ne sont plus nécessaires.

## Fichiers
Le contenu complet de `.github/workflows/mirador-autofix.yml` n'est pas disponible dans l'extrait de log ; le correctif est décrit ci-dessus plutôt que matérialisé automatiquement, pour éviter de produire un fichier de workflow incomplet ou incorrect.

## Détail du correctif

```
Remplacer le GITHUB_TOKEN par un token autorisé à créer des PR (GitHub App token via actions/create-github-app-token, ou PAT fine-grained stocké en secret) pour l'étape `gh pr create` du workflow .github/workflows/mirador-autofix.yml — OU activer le paramètre « Allow GitHub Actions to create and approve pull requests » dans Settings → Actions → General → Workflow permissions.
```

> Correctif à compléter/appliquer par un responsable (Mirador n'a pas pu le matérialiser automatiquement).
