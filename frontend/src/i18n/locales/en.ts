/**
 * en dictionary shim: upstream modular dictionary deep-merged with fork deltas.
 *
 * Background: this fork used to keep the full flat dictionary in this file,
 * which made `import('./locales/en')` resolve here and shadow the upstream
 * locales/en/ modular directory, so every translation key added upstream was
 * silently missing at runtime (mass [intlify] Not found warnings).
 *
 * Maintenance:
 * - upstream copy changes -> modules under locales/en/;
 * - fork-only copy / overrides -> locales/en-fork.ts (its values win);
 * - final runtime dictionary = deepMerge(locales/en/, locales/en-fork.ts).
 */
import { deepMerge } from '../deepMerge'

import modular from './en/index'
import fork from './en-fork'

export default deepMerge(modular, fork)
