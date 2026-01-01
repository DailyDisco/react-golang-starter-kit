import type { FuzzyMatch } from "@/services/command-palette/types";

/**
 * Simple fuzzy matching algorithm
 *
 * Scores based on:
 * - Consecutive character matches (bonus)
 * - Word boundary matches (bonus)
 * - Position in string (earlier = better)
 * - Case sensitivity (exact case = bonus)
 */
export function fuzzyMatch<T>(query: string, target: string, item: T): FuzzyMatch<T> | null {
  if (!query) {
    return { item, score: 0, matches: [] };
  }

  const queryLower = query.toLowerCase();
  const targetLower = target.toLowerCase();

  // Quick check: all query characters must exist in target
  let queryIndex = 0;
  for (let i = 0; i < targetLower.length && queryIndex < queryLower.length; i++) {
    if (targetLower[i] === queryLower[queryIndex]) {
      queryIndex++;
    }
  }

  if (queryIndex !== queryLower.length) {
    return null; // Not all characters found
  }

  // Calculate score and find matches
  let score = 0;
  const matches: Array<{ start: number; end: number }> = [];

  queryIndex = 0;
  let consecutiveMatches = 0;
  let currentMatchStart = -1;

  for (let i = 0; i < target.length && queryIndex < query.length; i++) {
    if (targetLower[i] === queryLower[queryIndex]) {
      // Start new match range
      if (currentMatchStart === -1) {
        currentMatchStart = i;
      }

      // Consecutive match bonus
      consecutiveMatches++;
      score += consecutiveMatches * 2;

      // Word boundary bonus (start of word)
      if (i === 0 || /\s|[_-]/.test(target[i - 1])) {
        score += 10;
      }

      // Exact case match bonus
      if (target[i] === query[queryIndex]) {
        score += 1;
      }

      // Earlier position bonus
      score += Math.max(0, 10 - i);

      queryIndex++;
    } else {
      // End current match range
      if (currentMatchStart !== -1) {
        matches.push({ start: currentMatchStart, end: i });
        currentMatchStart = -1;
      }
      consecutiveMatches = 0;
    }
  }

  // Close final match range
  if (currentMatchStart !== -1) {
    matches.push({ start: currentMatchStart, end: target.length });
  }

  // Bonus for exact match
  if (targetLower === queryLower) {
    score += 100;
  }

  // Bonus for prefix match
  if (targetLower.startsWith(queryLower)) {
    score += 50;
  }

  return { item, score, matches };
}

/**
 * Sort items by fuzzy match score
 */
export function fuzzySort<T>(items: T[], query: string, getText: (item: T) => string): Array<FuzzyMatch<T>> {
  if (!query) {
    return items.map((item) => ({ item, score: 0, matches: [] }));
  }

  const results: Array<FuzzyMatch<T>> = [];

  for (const item of items) {
    const text = getText(item);
    const match = fuzzyMatch(query, text, item);
    if (match) {
      results.push(match);
    }
  }

  // Sort by score descending
  return results.sort((a, b) => b.score - a.score);
}

/**
 * Highlight matched portions of text
 */
export function highlightMatches(
  text: string,
  matches: Array<{ start: number; end: number }>
): Array<{ text: string; highlighted: boolean }> {
  if (!matches.length) {
    return [{ text, highlighted: false }];
  }

  const result: Array<{ text: string; highlighted: boolean }> = [];
  let lastEnd = 0;

  for (const match of matches) {
    // Add non-highlighted portion before match
    if (match.start > lastEnd) {
      result.push({
        text: text.slice(lastEnd, match.start),
        highlighted: false,
      });
    }

    // Add highlighted match
    result.push({
      text: text.slice(match.start, match.end),
      highlighted: true,
    });

    lastEnd = match.end;
  }

  // Add remaining non-highlighted portion
  if (lastEnd < text.length) {
    result.push({
      text: text.slice(lastEnd),
      highlighted: false,
    });
  }

  return result;
}

/**
 * Filter and sort commands by fuzzy match
 */
export function filterByFuzzy<T extends { label: string; keywords?: string[] }>(items: T[], query: string): T[] {
  if (!query) return items;

  const results: Array<{ item: T; score: number }> = [];

  for (const item of items) {
    // Match against label
    const labelMatch = fuzzyMatch(query, item.label, item);

    // Match against keywords
    let keywordScore = 0;
    if (item.keywords) {
      for (const keyword of item.keywords) {
        const kwMatch = fuzzyMatch(query, keyword, item);
        if (kwMatch && kwMatch.score > keywordScore) {
          keywordScore = kwMatch.score;
        }
      }
    }

    const totalScore = (labelMatch?.score ?? 0) + keywordScore * 0.5;

    if (labelMatch || keywordScore > 0) {
      results.push({ item, score: totalScore });
    }
  }

  return results.sort((a, b) => b.score - a.score).map((r) => r.item);
}
