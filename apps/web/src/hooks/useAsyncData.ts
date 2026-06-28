"use client";

import { useEffect, useState } from "react";

/**
 * Fetches async data on mount. The fetch is wrapped in a
 * microtask via Promise.resolve().then(...) so no synchronous
 * setState happens inside the effect body. A cancelled flag
 * guards against setState-after-unmount.
 */
export function useAsyncData<T>(
  fetcher: () => Promise<T>,
  deps: ReadonlyArray<unknown> = []
): { data: T | null; loading: boolean; error: Error | null; reload: () => void } {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    Promise.resolve()
      .then(async () => {
        try {
          const result = await fetcher();
          if (cancelled) return;
          setData(result);
          setLoading(false);
        } catch (err: unknown) {
          if (cancelled) return;
          setError(err instanceof Error ? err : new Error(String(err)));
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reloadKey, ...deps]);

  return { data, loading, error, reload: () => setReloadKey((k) => k + 1) };
}
