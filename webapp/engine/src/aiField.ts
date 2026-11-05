// Reverse terrain-entry distance field, the same search as internal/dosai/field.go.
import { around } from "./hexgrid.ts";
import { coordKey } from "./level.ts";
import { coordFromKey } from "./playModel.ts";
import { at } from "./goruntime.ts";

interface Item {
  c: string;
  d: number;
}
export function dijkstra(
  costs: ReadonlyMap<string, number>,
  start: ReadonlyMap<string, number>,
): Map<string, number> {
  const dist = new Map(start),
    q: Item[] = [];
  const push = (it: Item): void => {
    q.push(it);
    let i = q.length - 1;
    while (i > 0) {
      const p = Math.trunc((i - 1) / 2);
      if (at(q, i).d >= at(q, p).d) break;
      [q[i], q[p]] = [at(q, p), at(q, i)];
      i = p;
    }
  };
  const pop = (): Item => {
    const out = at(q, 0),
      last = q.pop();
    if (q.length > 0 && last !== undefined) {
      q[0] = last;
      let i = 0;
      for (;;) {
        let j = 2 * i + 1;
        if (j >= q.length) break;
        if (j + 1 < q.length && at(q, j + 1).d < at(q, j).d) j++;
        if (at(q, j).d >= at(q, i).d) break;
        [q[i], q[j]] = [at(q, j), at(q, i)];
        i = j;
      }
    }
    return out;
  };
  for (const [c, d] of start) push({ c, d });
  while (q.length > 0) {
    const it = pop();
    if (it.d > (dist.get(it.c) ?? 0)) continue;
    const d = it.d + (costs.get(it.c) ?? 0);
    for (const c of around(coordFromKey(it.c))) {
      const n = coordKey(c);
      if (!costs.has(n)) continue;
      const old = dist.get(n);
      if (old === undefined || d < old) {
        dist.set(n, d);
        push({ c: n, d });
      }
    }
  }
  return dist;
}
