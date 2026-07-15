# Correctif proposé par Mirador

## Problème
Le workflow « Claude Code » (run 29431950707) échoue juste après l'installation de Claude Code :

```
Action failed with error: Environment variable validation failed:
  - Either ANTHROPIC_API_KEY, CLAUDE_CODE_OAUTH_TOKEN, or workload identity federation (ANTHROPIC_FEDERATION_RULE_ID and ANTHROPIC_ORGANIZATION_ID) is required when using direct Anthropic API.
```

Toutes les étapes préalables réussissent (checkout, setup-bun, bun install, obtention du token OIDC/App, création du commentaire). L'échec vient uniquement de l'absence d'identifiant Anthropic.

## Cause
Dans les logs, aucun secret d'authentification n'est transmis à l'action :

```
ANTHROPIC_API_KEY: 
CLAUDE_CODE_OAUTH_TOKEN: 
... "anthropic_api_key": "", "claude_code_oauth_token": "", ...
```

Le fichier `.github/workflows/claude.yml` appelle `anthropics/claude-code-action@v1` sans lui passer `anthropic_api_key` (ou `claude_code_oauth_token`). Ce n'est pas un incident transitoire : c'est une configuration manquante. Une relance échouera à l'identique.

## Correctif
1. **Ajouter le secret** dans le dépôt (Settings → Secrets and variables → Actions) :
   - soit `ANTHROPIC_API_KEY` (clé API Anthropic),
   - soit `CLAUDE_CODE_OAUTH_TOKEN` (jeton OAuth Claude Code).
2. **Câbler ce secret dans le workflow** `.github/workflows/claude.yml`, au niveau de l'étape qui utilise `anthropics/claude-code-action@v1` :

```yaml
      - uses: anthropics/claude-code-action@v1
        with:
          trigger_phrase: "@claude"
          label_trigger: claude
          # Fournir l'un des deux :
          anthropic_api_key: ${{ secrets.ANTHROPIC_API_KEY }}
          # ou bien :
          # claude_code_oauth_token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
```

> Variante entreprise : utiliser la fédération d'identité (`anthropic_federation_rule_id` + `anthropic_organization_id`) via OIDC si vous ne souhaitez pas stocker de clé longue durée.

## Vérification
Une fois le secret ajouté et le workflow mis à jour, relancer le run ou re-déclencher via un commentaire `@claude`. L'étape « Run Claude Code Action » ne doit plus lever l'erreur de validation d'environnement.

## Note
Le contenu complet de `.github/workflows/claude.yml` n'apparaît pas dans l'extrait de log ; le correctif est décrit ci-dessus plutôt que matérialisé automatiquement, afin d'éviter de produire un fichier de workflow incomplet ou incorrect. La modification d'un fichier sous `.github/workflows/` peut aussi requérir un token disposant de la permission `workflows: write`.

## Détail du correctif

```
Passer un secret d'authentification Anthropic (ANTHROPIC_API_KEY ou CLAUDE_CODE_OAUTH_TOKEN) à l'action anthropics/claude-code-action@v1 dans .github/workflows/claude.yml, et créer ce secret dans les paramètres du dépôt. L'erreur « Environment variable validation failed » vient de l'absence totale d'identifiant (tous les inputs d'auth sont vides dans les logs).
```

> Correctif à compléter/appliquer par un responsable (Mirador n'a pas pu le matérialiser automatiquement).
