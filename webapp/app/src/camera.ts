/** Native internal/camera transforms. Coordinates are physical screen pixels;
 * the Pixi adapter divides them by DPR when placing CSS-sized display objects. */
export interface Point {
  x: number;
  y: number;
}
export interface Size {
  width: number;
  height: number;
}
export interface Rect extends Point, Size {}

export class Camera {
  world: Size;
  view: Size;
  zoom = 1;
  scale = 1;
  origin: Point = { x: 0, y: 0 };
  constructor(world: Size, view: Size) {
    this.world = { ...world };
    this.view = { ...view };
    this.clamp();
  }
  effective(): number {
    return this.zoom * (this.scale <= 0 ? 1 : this.scale);
  }
  screenToWorld(x: number, y: number): Point {
    const e = this.effective();
    return { x: x / e + this.origin.x, y: y / e + this.origin.y };
  }
  worldToScreen(x: number, y: number): Point {
    const e = this.effective();
    return { x: (x - this.origin.x) * e, y: (y - this.origin.y) * e };
  }
  scroll(x: number, y: number): void {
    const e = this.effective();
    this.origin.x += x / e;
    this.origin.y += y / e;
    this.clamp();
  }
  zoomAt(factor: number, x: number, y: number): void {
    if (!Number.isFinite(factor) || factor <= 0) return;
    const point = this.screenToWorld(x, y);
    this.zoom = Math.min(4, Math.max(0.25, this.zoom * factor));
    this.origin = {
      x: point.x - x / this.effective(),
      y: point.y - y / this.effective(),
    };
    this.clamp();
  }
  resize(view: Size): void {
    const c = this.screenToWorld(this.view.width / 2, this.view.height / 2);
    this.view = { ...view };
    this.centerOn(c.x, c.y);
  }
  setScale(scale: number): void {
    const c = this.screenToWorld(this.view.width / 2, this.view.height / 2);
    this.scale = scale;
    this.centerOn(c.x, c.y);
  }
  viewport(cssWidth: number, cssHeight: number, dpr: number): void {
    const c = this.screenToWorld(this.view.width / 2, this.view.height / 2);
    this.scale = dpr;
    this.view = { width: cssWidth * dpr, height: cssHeight * dpr };
    this.centerOn(c.x, c.y);
  }
  centerOn(x: number, y: number): void {
    const e = this.effective();
    this.origin = {
      x: x - this.view.width / 2 / e,
      y: y - this.view.height / 2 / e,
    };
    this.clamp();
  }
  visible(): Rect {
    const e = this.effective();
    return {
      ...this.origin,
      width: this.view.width / e,
      height: this.view.height / e,
    };
  }
  private clamp(): void {
    const axis = (origin: number, world: number, view: number) =>
      world <= view
        ? (world - view) / 2
        : Math.min(Math.max(origin, 0), world - view);
    const e = this.effective();
    this.origin.x = axis(this.origin.x, this.world.width, this.view.width / e);
    this.origin.y = axis(
      this.origin.y,
      this.world.height,
      this.view.height / e,
    );
  }
}
