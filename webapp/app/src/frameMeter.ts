export const frameMeasure = "historyline.main-frame";
interface FrameClock {
  now(): number;
  measure(name: string, span: { start: number; end: number }): void;
  clearMeasures(name: string): void;
}

/** UserTiming observers receive complete map-loop/render spans. Clearing the
 * timeline entry immediately keeps diagnostics bounded during long games. */
export class FrameMeter {
  private start: number | null = null;
  constructor(private readonly clock: FrameClock = performance) {}
  begin(): void {
    this.start = this.clock.now();
  }
  end(): void {
    if (this.start === null) return;
    const start = this.start;
    this.start = null;
    this.clock.measure(frameMeasure, { start, end: this.clock.now() });
    this.clock.clearMeasures(frameMeasure);
  }
}
