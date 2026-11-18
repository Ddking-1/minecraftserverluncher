import { clamp, prefersReducedMotion } from '../core/dom.js';

/**
 * Self-contained Minecraft skin preview.
 *
 * The model is drawn with an orthographic projection and canvas 2D texture
 * sampling: every cuboid face is a parallelogram, so a single affine transform
 * per face maps the 64x64 skin region exactly. No three.js/CDN dependency,
 * which keeps Qt WebEngine offline-safe and compatible with the CSP.
 */

const HEAD_RECTS = {
  front: [8, 8, 8, 8],
  back: [24, 8, 8, 8],
  right: [0, 8, 8, 8],
  left: [16, 8, 8, 8],
  top: [8, 0, 8, 8],
  bottom: [16, 0, 8, 8],
};

const BODY_RECTS = {
  front: [4, 4, 8, 12],
  back: [16, 4, 8, 12],
  right: [0, 4, 4, 12],
  left: [12, 4, 4, 12],
  top: [4, 0, 8, 4],
  bottom: [12, 0, 8, 4],
};

const LIMB_RECTS = {
  front: [4, 4, 4, 12],
  back: [12, 4, 4, 12],
  right: [0, 4, 4, 12],
  left: [8, 4, 4, 12],
  top: [4, 0, 4, 4],
  bottom: [8, 0, 4, 4],
};

/**
 * Origin = top-left of the part inside the 64x64 texture.
 * overlay = origin of the second (hat/jacket) layer.
 */
const PARTS = [
  { id: 'head', w: 8, h: 8, d: 8, x: 0, y: -12, z: 0, origin: [0, 0], overlay: [32, 0], rects: HEAD_RECTS },
  { id: 'body', w: 8, h: 12, d: 4, x: 0, y: -2, z: 0, origin: [16, 16], overlay: [16, 32], rects: BODY_RECTS },
  { id: 'rightArm', w: 4, h: 12, d: 4, x: -6, y: -2, z: 0, origin: [40, 16], overlay: [40, 32], rects: LIMB_RECTS },
  { id: 'leftArm', w: 4, h: 12, d: 4, x: 6, y: -2, z: 0, origin: [32, 48], overlay: [48, 48], rects: LIMB_RECTS },
  { id: 'rightLeg', w: 4, h: 12, d: 4, x: -2, y: 10, z: 0, origin: [0, 16], overlay: [0, 32], rects: LIMB_RECTS },
  { id: 'leftLeg', w: 4, h: 12, d: 4, x: 2, y: 10, z: 0, origin: [16, 48], overlay: [0, 48], rects: LIMB_RECTS },
];

const FACE_SHADE = {
  front: 0.03,
  back: 0.15,
  right: 0.1,
  left: 0.07,
  top: 0,
  bottom: 0.22,
};

function rotatePoint(x, y, z, yaw, pitch) {
  const cy = Math.cos(yaw);
  const sy = Math.sin(yaw);
  const cp = Math.cos(pitch);
  const sp = Math.sin(pitch);
  const x1 = x * cy + z * sy;
  const z1 = -x * sy + z * cy;
  return {
    x: x1,
    y: y * cp - z1 * sp,
    z: y * sp + z1 * cp,
  };
}

function faceGeometry(w, h, d) {
  const hw = w / 2;
  const hh = h / 2;
  const hd = d / 2;
  return [
    { name: 'front', n: [0, 0, 1], p: [-hw, -hh, hd], u: [w, 0, 0], v: [0, h, 0] },
    { name: 'back', n: [0, 0, -1], p: [hw, -hh, -hd], u: [-w, 0, 0], v: [0, h, 0] },
    { name: 'right', n: [-1, 0, 0], p: [-hw, -hh, -hd], u: [0, 0, d], v: [0, h, 0] },
    { name: 'left', n: [1, 0, 0], p: [hw, -hh, hd], u: [0, 0, -d], v: [0, h, 0] },
    { name: 'top', n: [0, -1, 0], p: [-hw, -hh, -hd], u: [w, 0, 0], v: [0, 0, d] },
    { name: 'bottom', n: [0, 1, 0], p: [-hw, hh, hd], u: [w, 0, 0], v: [0, 0, -d] },
  ];
}

function copyMirrored(source, sx, sy, sw, sh, dest, dx, dy) {
  const context = dest.getContext('2d');
  context.save();
  context.imageSmoothingEnabled = false;
  context.translate(dx + sw, dy);
  context.scale(-1, 1);
  context.drawImage(source, sx, sy, sw, sh, 0, 0, sw, sh);
  context.restore();
}

export class SkinViewer {
  constructor(canvas, { skinUrl = '', autoRotate = true, showOverlay = true } = {}) {
    this.canvas = canvas;
    this.context = canvas.getContext('2d');
    this.skinUrl = '';
    this.texture = null;
    this.width = 0;
    this.height = 0;
    this.dpr = 1;
    this.yaw = 0.55;
    this.pitch = -0.12;
    this.autoRotate = autoRotate && !prefersReducedMotion();
    this.showOverlay = showOverlay;
    this.running = false;
    this.visible = true;
    this.dragging = false;
    this.resumeAt = 0;
    this.lastFrame = 0;
    this.frame = 0;
    this.destroyed = false;

    this.onPointerDown = this.onPointerDown.bind(this);
    this.onPointerMove = this.onPointerMove.bind(this);
    this.onPointerUp = this.onPointerUp.bind(this);
    this.onKeyDown = this.onKeyDown.bind(this);
    this.onVisibility = this.onVisibility.bind(this);
    this.loop = this.loop.bind(this);

    canvas.addEventListener('pointerdown', this.onPointerDown);
    canvas.addEventListener('pointermove', this.onPointerMove);
    canvas.addEventListener('pointerup', this.onPointerUp);
    canvas.addEventListener('pointercancel', this.onPointerUp);
    canvas.addEventListener('keydown', this.onKeyDown);
    document.addEventListener('visibilitychange', this.onVisibility);

    if ('ResizeObserver' in window) {
      this.resizeObserver = new ResizeObserver(() => this.resize());
      this.resizeObserver.observe(canvas);
    } else {
      this.onWindowResize = () => this.resize();
      window.addEventListener('resize', this.onWindowResize);
    }

    if ('IntersectionObserver' in window) {
      this.intersectionObserver = new IntersectionObserver(
        (entries) => {
          this.visible = entries.some((entry) => entry.isIntersecting);
        },
        { threshold: 0.01 },
      );
      this.intersectionObserver.observe(canvas);
    }

    this.resize();
    if (skinUrl) this.setSkin(skinUrl).catch((error) => console.warn('skin preview failed', error));
    this.start();
  }

  async setSkin(url) {
    if (!url) throw new Error('Hiányzó skin.');
    const image = await new Promise((resolve, reject) => {
      const img = new Image();
      img.decoding = 'async';
      img.onload = () => resolve(img);
      img.onerror = () => reject(new Error('A skin kép nem tölthető be.'));
      img.src = url;
    });

    const valid =
      (image.naturalWidth === 64 && image.naturalHeight === 64) ||
      (image.naturalWidth === 64 && image.naturalHeight === 32);
    if (!valid) throw new Error('A skin 64×64 vagy 64×32 pixeles PNG legyen.');

    const texture = document.createElement('canvas');
    texture.width = 64;
    texture.height = 64;
    const context = texture.getContext('2d');
    context.imageSmoothingEnabled = false;
    context.drawImage(image, 0, 0);

    // Legacy 64x32 skins have no left-limb texture; mirror the right limbs.
    if (image.naturalHeight === 32) {
      copyMirrored(image, 40, 16, 16, 16, texture, 32, 48);
      copyMirrored(image, 0, 16, 16, 16, texture, 16, 48);
    }

    this.texture = texture;
    this.skinUrl = url;
    this.model = image.naturalHeight === 64 ? '64×64 · 1.8' : '64×32 · Legacy';
    this.render();
    return { width: image.naturalWidth, height: image.naturalHeight, model: this.model };
  }

  resize() {
    if (this.destroyed) return;
    const rect = this.canvas.getBoundingClientRect();
    const width = Math.max(1, Math.round(rect.width));
    const height = Math.max(1, Math.round(rect.height));
    const dpr = clamp(window.devicePixelRatio || 1, 1, 2);

    if (width === this.width && height === this.height && dpr === this.dpr) return;
    this.width = width;
    this.height = height;
    this.dpr = dpr;
    this.canvas.width = Math.round(width * dpr);
    this.canvas.height = Math.round(height * dpr);
    this.render();
  }

  start() {
    if (this.running || this.destroyed) return;
    this.running = true;
    this.lastFrame = 0;
    this.frame = window.requestAnimationFrame(this.loop);
  }

  stop() {
    this.running = false;
    window.cancelAnimationFrame(this.frame);
  }

  loop(time) {
    if (!this.running || this.destroyed) return;
    const delta = this.lastFrame ? time - this.lastFrame : 16;
    this.lastFrame = time;

    if (this.autoRotate && !this.dragging && this.visible && time > this.resumeAt) {
      this.yaw += delta * 0.00028;
    }

    this.render();
    this.frame = window.requestAnimationFrame(this.loop);
  }

  onVisibility() {
    if (document.hidden) this.stop();
    else this.start();
  }

  onPointerDown(event) {
    if (event.button !== undefined && event.button !== 0) return;
    this.dragging = true;
    this.startX = event.clientX;
    this.startY = event.clientY;
    this.startYaw = this.yaw;
    this.startPitch = this.pitch;
    try {
      this.canvas.setPointerCapture?.(event.pointerId);
    } catch (error) {
      /* Synthetic or already-released pointers do not need capture. */
    }
    this.canvas.classList.add('is-dragging');
    event.preventDefault();
  }

  onPointerMove(event) {
    if (!this.dragging) return;
    const dx = event.clientX - this.startX;
    const dy = event.clientY - this.startY;
    this.yaw = this.startYaw + dx * 0.012;
    this.pitch = clamp(this.startPitch + dy * 0.008, -0.6, 0.6);
    this.render();
  }

  onPointerUp(event) {
    if (!this.dragging) return;
    this.dragging = false;
    this.resumeAt = performance.now() + 2600;
    try {
      this.canvas.releasePointerCapture?.(event.pointerId);
    } catch (error) {
      /* The pointer may already have been released by the browser. */
    }
    this.canvas.classList.remove('is-dragging');
  }

  onKeyDown(event) {
    const step = event.shiftKey ? 0.25 : 0.08;
    if (event.key === 'ArrowLeft') this.yaw -= step;
    else if (event.key === 'ArrowRight') this.yaw += step;
    else if (event.key === 'ArrowUp') this.pitch = clamp(this.pitch - step, -0.6, 0.6);
    else if (event.key === 'ArrowDown') this.pitch = clamp(this.pitch + step, -0.6, 0.6);
    else if (event.key === 'r' || event.key === 'R') this.resetView();
    else return;
    event.preventDefault();
    this.resumeAt = performance.now() + 2600;
    this.render();
  }

  resetView() {
    this.yaw = 0.55;
    this.pitch = -0.12;
    this.resumeAt = performance.now() + 1600;
    this.render();
  }

  setOverlay(enabled) {
    this.showOverlay = Boolean(enabled);
    this.render();
  }

  setAutoRotate(enabled) {
    this.autoRotate = Boolean(enabled) && !prefersReducedMotion();
  }

  render() {
    if (!this.texture || !this.width || !this.height) return;
    const ctx = this.context;
    const dpr = this.dpr;
    const width = this.width * dpr;
    const height = this.height * dpr;
    const unit = Math.min(width / 24, height / 40);
    const centerX = width / 2;
    const centerY = height / 2 + unit * 0.4;

    ctx.setTransform(1, 0, 0, 1, 0, 0);
    ctx.clearRect(0, 0, width, height);

    const project = (x, y, z) => {
      const point = rotatePoint(x, y, z, this.yaw, this.pitch);
      return { x: centerX + point.x * unit, y: centerY + point.y * unit, z: point.z };
    };

    this.drawShadow(ctx, centerX, centerY, unit);

    const faces = [];
    for (const part of PARTS) {
      const layers = [{ inflate: 0, overlay: false, origin: part.origin }];
      if (this.showOverlay) layers.push({ inflate: 0.55, overlay: true, origin: part.overlay });

      for (const layer of layers) {
        const { w, h, d } = part;
        const iw = w + layer.inflate;
        const ih = h + layer.inflate;
        const id = d + layer.inflate;

        for (const face of faceGeometry(iw, ih, id)) {
          const rect = part.rects[face.name];
          if (!rect) continue;
          const topLeft = {
            x: part.x + face.p[0],
            y: part.y + face.p[1],
            z: part.z + face.p[2],
          };
          const center = {
            x: topLeft.x + face.u[0] / 2 + face.v[0] / 2,
            y: topLeft.y + face.u[1] / 2 + face.v[1] / 2,
            z: topLeft.z + face.u[2] / 2 + face.v[2] / 2,
          };
          const rotatedCenter = rotatePoint(center.x, center.y, center.z, this.yaw, this.pitch);
          const normal = rotatePoint(face.n[0], face.n[1], face.n[2], this.yaw, this.pitch);
          if (normal.z <= 0.001) continue;

          faces.push({
            ...face,
            rect,
            origin: layer.origin,
            overlay: layer.overlay,
            topLeft,
            depth: rotatedCenter.z,
            shade: FACE_SHADE[face.name] + (layer.overlay ? 0.03 : 0),
          });
        }
      }
    }

    faces.sort((a, b) => a.depth - b.depth);

    for (const face of faces) {
      const p0 = project(face.topLeft.x, face.topLeft.y, face.topLeft.z);
      const p1 = project(
        face.topLeft.x + face.u[0],
        face.topLeft.y + face.u[1],
        face.topLeft.z + face.u[2],
      );
      const p3 = project(
        face.topLeft.x + face.v[0],
        face.topLeft.y + face.v[1],
        face.topLeft.z + face.v[2],
      );

      const [sx, sy, sw, sh] = face.rect;
      const tx = face.origin[0] + sx;
      const ty = face.origin[1] + sy;

      const a = (p1.x - p0.x) / sw;
      const b = (p1.y - p0.y) / sw;
      const c = (p3.x - p0.x) / sh;
      const d = (p3.y - p0.y) / sh;

      ctx.save();
      ctx.beginPath();
      ctx.moveTo(p0.x, p0.y);
      ctx.lineTo(p1.x, p1.y);
      ctx.lineTo(p1.x + (p3.x - p0.x), p1.y + (p3.y - p0.y));
      ctx.lineTo(p3.x, p3.y);
      ctx.closePath();
      ctx.clip();
      ctx.setTransform(a, b, c, d, p0.x, p0.y);
      ctx.imageSmoothingEnabled = false;
      ctx.drawImage(this.texture, tx, ty, sw, sh, 0, 0, sw, sh);
      ctx.restore();

      // The jacket/hat layer is mostly transparent; shading the whole quad would
      // paint a dark shell over the base model. Base faces are opaque, so only
      // those receive the directional tint.
      if (face.shade > 0.001 && !face.overlay) {
        ctx.save();
        ctx.setTransform(1, 0, 0, 1, 0, 0);
        ctx.beginPath();
        ctx.moveTo(p0.x, p0.y);
        ctx.lineTo(p1.x, p1.y);
        ctx.lineTo(p1.x + (p3.x - p0.x), p1.y + (p3.y - p0.y));
        ctx.lineTo(p3.x, p3.y);
        ctx.closePath();
        ctx.fillStyle = `rgba(3, 7, 12, ${face.shade})`;
        ctx.fill();
        ctx.restore();
      }
    }
  }

  drawShadow(ctx, centerX, centerY, unit) {
    const foot = { x: centerX, y: centerY + 16.6 * unit };
    const gradient = ctx.createRadialGradient(foot.x, foot.y, 0, foot.x, foot.y, unit * 9);
    gradient.addColorStop(0, 'rgba(2, 6, 12, 0.34)');
    gradient.addColorStop(1, 'rgba(2, 6, 12, 0)');
    ctx.save();
    ctx.setTransform(1, 0, 0, 1, 0, 0);
    ctx.fillStyle = gradient;
    ctx.beginPath();
    ctx.ellipse(foot.x, foot.y, unit * 9, unit * 2.6, 0, 0, Math.PI * 2);
    ctx.fill();
    ctx.restore();
  }

  destroy() {
    this.destroyed = true;
    this.stop();
    this.canvas.removeEventListener('pointerdown', this.onPointerDown);
    this.canvas.removeEventListener('pointermove', this.onPointerMove);
    this.canvas.removeEventListener('pointerup', this.onPointerUp);
    this.canvas.removeEventListener('pointercancel', this.onPointerUp);
    this.canvas.removeEventListener('keydown', this.onKeyDown);
    document.removeEventListener('visibilitychange', this.onVisibility);
    if (this.resizeObserver) this.resizeObserver.disconnect();
    if (this.intersectionObserver) this.intersectionObserver.disconnect();
    if (this.onWindowResize) window.removeEventListener('resize', this.onWindowResize);
  }
}

export default SkinViewer;
