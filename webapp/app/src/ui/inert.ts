interface Lock {
  count: number;
  previous: boolean;
}
const locks = new WeakMap<HTMLElement, Lock>();

/** Modal lifetimes can overlap while Preact replaces one page with another. */
export function makeInert(node: HTMLElement): () => void {
  let lock = locks.get(node);
  if (lock === undefined) {
    lock = { count: 0, previous: node.inert };
    locks.set(node, lock);
  }
  lock.count++;
  node.inert = true;
  let released = false;
  return () => {
    if (released) return;
    released = true;
    lock.count--;
    if (lock.count === 0) {
      node.inert = lock.previous;
      locks.delete(node);
    }
  };
}
