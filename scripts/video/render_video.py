#!/usr/bin/env python3
"""Render the mr-ducksafe promo videos and thumbnail.

Outputs (in docs/media/):
  mr-ducksafe-youtube.mp4   1920x1080, ~90 s  (YouTube)
  mr-ducksafe-short.mp4     1080x1920, ~35 s  (YouTube Shorts / Reels / TikTok)
  thumbnail.png             1280x720          (YouTube thumbnail, README)

Every frame is drawn with Pillow and piped to ffmpeg; the soundtrack
(pad + bass + drums + keyboard clicks) is synthesized with numpy. Terminal
content comes from the real captures in scripts/demo/*.txt.

Usage:
    python scripts/video/render_video.py                 # everything
    python scripts/video/render_video.py thumbnail       # youtube | short | thumbnail
    python scripts/video/render_video.py preview youtube # a few still frames only

Requires Python 3.10+, Pillow, numpy and ffmpeg on PATH.
"""
from __future__ import annotations

import math
import multiprocessing as mp
import os
import re
import subprocess
import sys
import tempfile
import wave
from functools import lru_cache
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw, ImageFilter, ImageFont

ROOT = Path(__file__).resolve().parents[2]
DEMO = ROOT / "scripts" / "demo"
OUT = ROOT / "docs" / "media"
FPS = 30
SR = 44100

# --------------------------------------------------------------------------
# Palette (Catppuccin Mocha, same as docs/demo.gif)
# --------------------------------------------------------------------------
CRUST = (17, 17, 27)
MANTLE = (24, 24, 37)
BASE = (30, 30, 46)
SURF0 = (49, 50, 68)
SURF1 = (69, 71, 90)
OVER0 = (108, 112, 134)
SUB = (166, 173, 200)
TEXT = (205, 214, 244)
WHITE = (240, 243, 255)
RED = (243, 139, 168)
PEACH = (250, 179, 135)
YELLOW = (249, 226, 175)
GREEN = (166, 227, 161)
TEAL = (148, 226, 213)
SKY = (137, 220, 235)
BLUE = (137, 180, 250)
MAUVE = (203, 166, 247)
DUCK = (255, 204, 77)
CREAM = (245, 224, 220)

SEV_COLOR = {"KRIT": RED, "YÜKS": PEACH, "ORTA": YELLOW, "DÜŞK": OVER0}

# --------------------------------------------------------------------------
# Fonts
# --------------------------------------------------------------------------
FONT_PATHS = {
    "black": ["/usr/share/fonts/opentype/inter/InterDisplay-Black.otf",
              r"C:\Windows\Fonts\seguibl.ttf"],
    "xbold": ["/usr/share/fonts/opentype/inter/InterDisplay-ExtraBold.otf",
              r"C:\Windows\Fonts\segoeuib.ttf"],
    "bold": ["/usr/share/fonts/opentype/inter/InterDisplay-Bold.otf",
             r"C:\Windows\Fonts\segoeuib.ttf"],
    "semi": ["/usr/share/fonts/opentype/inter/Inter-SemiBold.otf",
             r"C:\Windows\Fonts\seguisb.ttf"],
    "med": ["/usr/share/fonts/opentype/inter/Inter-Medium.otf",
            r"C:\Windows\Fonts\segoeui.ttf"],
    "reg": ["/usr/share/fonts/opentype/inter/Inter-Regular.otf",
            r"C:\Windows\Fonts\segoeui.ttf"],
    "mono": ["/usr/share/fonts/truetype/jetbrains-mono/JetBrainsMono-Regular.ttf",
             r"C:\Windows\Fonts\CascadiaMono.ttf",
             "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"],
    "monob": ["/usr/share/fonts/truetype/jetbrains-mono/JetBrainsMono-Bold.ttf",
              r"C:\Windows\Fonts\CascadiaMono-Bold.ttf",
              "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Bold.ttf"],
    "emoji": ["/usr/share/fonts/truetype/noto/NotoColorEmoji.ttf",
              r"C:\Windows\Fonts\seguiemj.ttf"],
}


@lru_cache(maxsize=None)
def font(key: str, size: int) -> ImageFont.FreeTypeFont:
    for p in FONT_PATHS[key]:
        if os.path.exists(p):
            return ImageFont.truetype(p, size)
    raise SystemExit(f"font not found for {key!r}: {FONT_PATHS[key]}")


# --------------------------------------------------------------------------
# Math helpers
# --------------------------------------------------------------------------
def clamp(x, a=0.0, b=1.0):
    return a if x < a else b if x > b else x


def ease_out(x):
    x = clamp(x)
    return 1 - (1 - x) ** 3


def ease_in_out(x):
    x = clamp(x)
    return 4 * x ** 3 if x < 0.5 else 1 - (-2 * x + 2) ** 3 / 2


def ease_back(x):
    x = clamp(x)
    c1, c3 = 1.70158, 2.70158
    return 1 + c3 * (x - 1) ** 3 + c1 * (x - 1) ** 2


def ap(t, start, dur=0.45):
    """Appear progress 0..1 (eased) for an element that starts at `start`."""
    return ease_out((t - start) / dur)


def mix(c1, c2, a):
    return tuple(int(c1[i] + (c2[i] - c1[i]) * a) for i in range(3))


# --------------------------------------------------------------------------
# Sprites (cached RGBA images) + compositing
# --------------------------------------------------------------------------
@lru_cache(maxsize=8192)
def text_sprite(s: str, key: str, size: int, color, stroke=0, stroke_color=CRUST):
    f = font(key, size)
    asc, desc = f.getmetrics()
    w = int(math.ceil(f.getlength(s))) + 2 * stroke + 2
    h = asc + desc + 2 * stroke
    img = Image.new("RGBA", (max(w, 1), max(h, 1)), (0, 0, 0, 0))
    ImageDraw.Draw(img).text((stroke, stroke), s, font=f, fill=color + (255,),
                             stroke_width=stroke, stroke_fill=stroke_color + (255,))
    return img


@lru_cache(maxsize=256)
def emoji_sprite(ch: str, size: int):
    ch = ch.replace("\ufe0f", "")
    f = font("emoji", 109)
    img = Image.new("RGBA", (160, 160), (0, 0, 0, 0))
    ImageDraw.Draw(img).text((8, 8), ch, font=f, embedded_color=True)
    bbox = img.getbbox() or (0, 0, 1, 1)
    img = img.crop(bbox)
    scale = size / max(img.size)
    return img.resize((max(1, int(img.width * scale)), max(1, int(img.height * scale))),
                      Image.LANCZOS)


@lru_cache(maxsize=512)
def panel_sprite(w, h, r, fill, outline=None, width=2, fill_alpha=255):
    img = Image.new("RGBA", (w, h), (0, 0, 0, 0))
    ImageDraw.Draw(img).rounded_rectangle(
        [0, 0, w - 1, h - 1], radius=r, fill=fill + (fill_alpha,),
        outline=(outline + (255,)) if outline else None, width=width)
    return img


@lru_cache(maxsize=256)
def shadow_sprite(w, h, r, blur=28, strength=170):
    pad = blur * 2
    img = Image.new("RGBA", (w + 2 * pad, h + 2 * pad), (0, 0, 0, 0))
    ImageDraw.Draw(img).rounded_rectangle([pad, pad + blur // 3, pad + w, pad + h + blur // 3],
                                          radius=r, fill=(0, 0, 0, strength))
    return img.filter(ImageFilter.GaussianBlur(blur))


@lru_cache(maxsize=128)
def glow_sprite(radius, color, strength=120):
    size = radius * 2
    y, x = np.ogrid[:size, :size]
    d = np.sqrt((x - radius) ** 2 + (y - radius) ** 2) / radius
    a = np.clip(1 - d, 0, 1) ** 2 * strength
    arr = np.zeros((size, size, 4), np.uint8)
    arr[..., 0], arr[..., 1], arr[..., 2] = color
    arr[..., 3] = a.astype(np.uint8)
    return Image.fromarray(arr, "RGBA")


_ALPHA_LUT = {}


def put(canvas, sprite, x, y, anchor="lt", alpha=1.0, scale=1.0):
    """Paste sprite at (x, y). anchor: l/m/r + t/m/b."""
    if alpha <= 0.004:
        return
    if scale != 1.0:
        if scale <= 0.01:
            return
        sprite = sprite.resize((max(1, int(sprite.width * scale)),
                                max(1, int(sprite.height * scale))), Image.BILINEAR)
    w, h = sprite.size
    ax, ay = anchor[0], anchor[1]
    if ax == "m":
        x -= w / 2
    elif ax == "r":
        x -= w
    if ay == "m":
        y -= h / 2
    elif ay == "b":
        y -= h
    mask = sprite.getchannel("A")
    if alpha < 0.996:
        q = int(alpha * 100)
        lut = _ALPHA_LUT.get(q)
        if lut is None:
            lut = _ALPHA_LUT[q] = [int(v * q / 100) for v in range(256)]
        mask = mask.point(lut)
    canvas.paste(sprite, (int(round(x)), int(round(y))), mask)


def txt(canvas, s, x, y, key, size, color, anchor="lt", alpha=1.0, scale=1.0, stroke=0):
    put(canvas, text_sprite(s, key, size, color, stroke), x, y, anchor, alpha, scale)


def tw(s, key, size):
    return font(key, size).getlength(s)


def rich(canvas, parts, x, y, size, anchor="lt", alpha=1.0, key="bold"):
    """Draw [(text, color[, fontkey])] parts on one line."""
    total = sum(tw(p[0], p[2] if len(p) > 2 else key, size) for p in parts)
    if anchor[0] == "m":
        x -= total / 2
    elif anchor[0] == "r":
        x -= total
    for p in parts:
        k = p[2] if len(p) > 2 else key
        txt(canvas, p[0], x, y, k, size, p[1], "l" + anchor[1], alpha)
        x += tw(p[0], k, size)
    return total


def card(canvas, x, y, w, h, r=22, fill=MANTLE, outline=SURF0, alpha=1.0, shadow=True):
    w, h = int(w), int(h)
    if shadow:
        sh = shadow_sprite(w, h, r)
        put(canvas, sh, x - (sh.width - w) / 2, y - (sh.height - h) / 2, alpha=alpha)
    put(canvas, panel_sprite(w, h, r, fill, outline), x, y, alpha=alpha)


def pill(canvas, s, x, y, size, fg, bg, outline=None, anchor="mt", alpha=1.0, key="semi",
         padx=None, pady=None):
    padx = padx if padx is not None else int(size * 0.75)
    pady = pady if pady is not None else int(size * 0.32)
    w = int(tw(s, key, size)) + 2 * padx
    h = int(sum(font(key, size).getmetrics())) + 2 * pady
    if anchor[0] == "m":
        x -= w / 2
    elif anchor[0] == "r":
        x -= w
    if anchor[1] == "m":
        y -= h / 2
    put(canvas, panel_sprite(w, h, h // 2, bg, outline), x, y, alpha=alpha)
    txt(canvas, s, x + padx, y + pady, key, size, fg, alpha=alpha)
    return w, h


# --------------------------------------------------------------------------
# Backgrounds
# --------------------------------------------------------------------------
@lru_cache(maxsize=4)
def background(w, h):
    y = np.linspace(0, 1, h)[:, None]
    x = np.linspace(0, 1, w)[None, :]
    img = np.zeros((h, w, 3))
    for c in range(3):
        img[..., c] = BASE[c] * (1 - y) + CRUST[c] * y
    aspect = w / h

    def glow(cx, cy, rad, col, k):
        d = np.sqrt(((x - cx) * aspect) ** 2 + (y - cy) ** 2) / rad
        g = np.clip(1 - d, 0, 1) ** 2 * k
        for c in range(3):
            img[..., c] += (col[c] - img[..., c]) * g

    glow(0.05, 0.0, 0.9, MAUVE, 0.10)
    glow(0.95, 1.0, 0.9, TEAL, 0.07)
    # dot grid
    step = 40 if w > h else 36
    out = Image.fromarray(np.clip(img, 0, 255).astype(np.uint8), "RGB")
    d = ImageDraw.Draw(out)
    for gy in range(step // 2, h, step):
        for gx in range(step // 2, w, step):
            d.point((gx, gy), fill=(52, 54, 74))
    return out


# --------------------------------------------------------------------------
# Terminal content: colorize captured lines into segments
# --------------------------------------------------------------------------
# A segment is (text, color, style) where style is "" | "b" | "badge".
def S(text, color=TEXT, style=""):
    return (text, color, style)


SEV_RE = re.compile(r"^(\s*)(KRIT|YÜKS|ORTA|DÜŞK)(\s+)(.+?)(\s{2,})(\S.*)$")
COUNT_WORDS = {"kritik": RED, "yüksek": PEACH, "orta": YELLOW, "düşük": OVER0}


def colorize(line: str):
    line = line.replace("\ufeff", "")
    s = line.strip()
    if not s:
        return [S("")]
    m = SEV_RE.match(line)
    if m and not s.startswith("│"):
        lead, sev, sp, app, sp2, ver = m.groups()
        return [S(lead), S(sev, SEV_COLOR[sev], "badge"), S(sp), S(app, WHITE, "b"),
                S(sp2), S(ver, OVER0)]
    lead = line[: len(line) - len(line.lstrip())]
    if re.match(r"^\d+ (kritik|yüksek|orta|düşük)", s):
        segs = [S(lead)]
        for i, part in enumerate(s.split(" · ")):
            if i:
                segs.append(S(" · ", OVER0))
            word = part.split(" ", 1)[1]
            segs.append(S(part, COUNT_WORDS.get(word, TEXT)))
        return segs
    for label, col in (("Aciliyet:", None), ("Risk:", TEXT), ("Çözüm:", TEAL), ("Detay:", OVER0)):
        if s.startswith(label):
            rest = line[len(lead) + len(label):]
            if col is None:
                col = RED if "Acilen" in rest else PEACH
            style = "b" if label == "Çözüm:" and "winget" in rest else ""
            return [S(lead), S(label, OVER0), S(rest, col, style)]
    if s.startswith("mr-ducksafe —") or s.startswith("mr-ducksafe scan #"):
        return [S(lead), S(s, MAUVE, "b")]
    if s.startswith("Bulgular:"):
        segs = [S("Bulgular:", MAUVE, "b")]
        for name, col in (("KRITIK", RED), ("YÜKSEK", PEACH), ("ORTA", YELLOW), ("DÜŞÜK", OVER0)):
            mm = re.search(name + r" (\d+)", s)
            if mm:
                segs.append(S(f"  {name} {mm.group(1)}", col))
        return segs
    if s[0] in "┌├└":
        return [S(line, SURF1)]
    if s.startswith("│"):
        cells = line.split("│")
        segs = []
        for i, c in enumerate(cells):
            if i:
                segs.append(S("│", SURF1))
            cs = c.strip()
            if cs in SEV_COLOR:
                segs.append(S(c, SEV_COLOR[cs], "b"))
            elif cs.startswith("CVE-"):
                segs.append(S(c, SKY))
            elif cs in ("SEV", "UYGULAMA", "SÜRÜM", "CVE", "DÜZELTME", "ÖZET"):
                segs.append(S(c, MAUVE, "b"))
            elif i == 2:
                segs.append(S(c, WHITE))
            else:
                segs.append(S(c, SUB if i == 6 else OVER0))
        return segs
    if re.match(r"^\d+\s+\d{4}-", s):
        segs = []
        for tok in re.split(r"(\s+)", line):
            col = {"success": GREEN, "running": YELLOW, "failed": RED}.get(tok, TEXT)
            if re.match(r"^\d{4}-|^\d\d:\d\d", tok):
                col = SUB
            segs.append(S(tok, col))
        return segs
    if s.startswith("ID "):
        return [S(line, MAUVE, "b")]
    return [S(line)]


def load_capture(name):
    lines = (DEMO / name).read_text(encoding="utf-8-sig").splitlines()
    while lines and not lines[-1].strip():
        lines.pop()
    return [colorize(ln) for ln in lines]


BANNER = [
    [S("")],
    [S("  __", DUCK)],
    [S(" <(o )___", DUCK), S("   "), S("mr-ducksafe", SKY, "b"), S(" v0.1.0", OVER0)],
    [S("  ( ._> /", DUCK), S("   "), S("yerel kurulu yazilim CVE tarayici", OVER0)],
    [S("   `---'", DUCK)],
]
PROMPT = "PS C:\\> "
SPIN = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"


def done(msg):
    return [S("✓", GREEN, "b"), S(" " + msg)]


class TermScript:
    """A timeline of terminal events → state at time t."""

    def __init__(self, prompt=PROMPT):
        self.prompt = prompt
        self.events = []
        self.t = 0.0

    def _add(self, ev, dur):
        ev["start"], ev["dur"] = self.t, dur
        self.events.append(ev)
        self.t += dur
        return ev["start"] + dur

    def type(self, cmd, cps=24, pre=0.5, post=0.3):
        return self._add({"k": "type", "cmd": cmd, "cps": cps, "pre": pre},
                         pre + len(cmd) / cps + post)

    def out(self, lines, rate=None):
        dur = 0.0 if rate is None else len(lines) / rate
        return self._add({"k": "out", "lines": lines, "rate": rate}, dur)

    def spin(self, label, dur, result):
        return self._add({"k": "spin", "label": label, "result": result}, dur)

    def count(self, fmt, n, dur, result):
        return self._add({"k": "count", "fmt": fmt, "n": n, "result": result}, dur)

    def wait(self, dur):
        return self._add({"k": "wait"}, dur)

    @property
    def duration(self):
        return self.t

    def clicks(self):
        out = []
        for ev in self.events:
            if ev["k"] == "type":
                for i in range(len(ev["cmd"])):
                    out.append(ev["start"] + ev["pre"] + i / ev["cps"])
        return out

    def prompt_line(self, cmd=""):
        return [S(self.prompt, SKY, "b"), S(cmd, CREAM)]

    def state(self, t):
        lines, partial, cursor = [], None, False
        for ev in self.events:
            if t < ev["start"]:
                break
            el = t - ev["start"]
            fin = el >= ev["dur"]
            k = ev["k"]
            if k == "type":
                if fin:
                    lines.append(self.prompt_line(ev["cmd"]))
                    partial = None
                else:
                    n = int(clamp((el - ev["pre"]) * ev["cps"], 0, len(ev["cmd"])))
                    partial = self.prompt_line(ev["cmd"][:n])
                    typing = 0 < n < len(ev["cmd"])
                    cursor = typing or int(t * 2.4) % 2 == 0
            elif k == "out":
                if ev["rate"] is None or fin:
                    lines.extend(ev["lines"])
                else:
                    lines.extend(ev["lines"][: int(el * ev["rate"]) + 1])
                partial = None
            elif k in ("spin", "count"):
                if fin:
                    lines.append(ev["result"])
                    partial = None
                else:
                    g = SPIN[int(t * 12) % len(SPIN)]
                    label = ev["label"] if k == "spin" else \
                        ev["fmt"].format(int(ev["n"] * clamp(el / ev["dur"])))
                    partial = [S(g, SKY), S(" " + label, SUB)]
                    cursor = False
        else:
            if t >= self.t:
                partial = self.prompt_line("")
                cursor = int(t * 2.4) % 2 == 0
        return lines, partial, cursor


def draw_terminal(canvas, box, script: TermScript | None, t, size, title="Windows PowerShell",
                  static_lines=None, alpha=1.0):
    x0, y0, x1, y1 = (int(v) for v in box)
    w, h = x1 - x0, y1 - y0
    card(canvas, x0, y0, w, h, r=16, fill=(21, 21, 33), outline=SURF0, alpha=alpha)
    bar = int(size * 2.1)
    put(canvas, panel_sprite(w - 2, bar, 15, CRUST), x0 + 1, y0 + 1, alpha=alpha)
    put(canvas, panel_sprite(w - 2, bar // 2, 0, CRUST), x0 + 1, y0 + 1 + bar // 2, alpha=alpha)
    d = ImageDraw.Draw(canvas)
    rr = max(5, size // 3)
    for i, c in enumerate((RED, YELLOW, GREEN)):
        cx, cy = x0 + 22 + i * (rr * 2 + 10) + rr, y0 + bar / 2
        d.ellipse([cx - rr, cy - rr, cx + rr, cy + rr], fill=mix(CRUST, c, alpha))
    txt(canvas, title, x0 + w / 2, y0 + bar / 2, "med", int(size * 0.8), OVER0, "mm", alpha)

    if script is not None:
        lines, partial, cursor = script.state(t)
    else:
        lines, partial, cursor = static_lines or [], None, False
    vis = list(lines) + ([partial] if partial is not None else [])
    row_h = int(size * 1.5)
    pad = int(size * 1.0)
    rows = (h - bar - 2 * pad) // row_h
    vis = vis[-rows:]
    f, fb = font("mono", size), font("monob", size)
    cw = f.getlength("M")
    asc = f.getmetrics()[0]
    for ri, segs in enumerate(vis):
        y = y0 + bar + pad + ri * row_h
        col = 0
        for text, color, style in segs:
            if not text:
                continue
            x = x0 + pad + col * cw
            color = mix((21, 21, 33), color, alpha)
            if style == "badge":
                bx0, bx1 = x - cw * 0.45, x + len(text) * cw + cw * 0.45
                d.rounded_rectangle([bx0, y - row_h * 0.08, bx1, y + row_h * 0.86],
                                    radius=int(size * 0.3), fill=color)
                d.text((x, y), text, font=fb, fill=CRUST)
            else:
                d.text((x, y), text, font=fb if style == "b" else f, fill=color)
            col += len(text)
        if cursor and ri == len(vis) - 1 and partial is not None:
            x = x0 + pad + col * cw
            d.rectangle([x + 1, y + 2, x + cw - 1, y + asc + 3], fill=mix((21, 21, 33), TEXT, alpha))


# --------------------------------------------------------------------------
# Scene framework
# --------------------------------------------------------------------------
class Scene:
    def __init__(self, dur, fn, sfx=None, name="", chapter=None):
        self.dur, self.fn, self.name, self.chapter = dur, fn, name, chapter
        self.sfx = sfx or []  # [(local_time, kind)]


def caption(canvas, W, t, parts, y=54, size=44):
    a = ap(t, 0.15, 0.5)
    rich(canvas, parts, W / 2, y + (1 - a) * -20, size, "mt", a)


# ======================= YouTube (16:9) scenes =============================
def yt_scenes(W=1920, H=1080):
    scenes = []
    cx = W / 2

    # 1 — Hook ---------------------------------------------------------------
    def hook(c, t):
        a1, a2 = ap(t, 0.2), ap(t, 0.7)
        txt(c, "Sıradan bir Windows bilgisayarında", cx, 250 + (1 - a1) * 30, "semi", 52, SUB, "mt", a1)
        txt(c, "kaç bilinen güvenlik açığı var?", cx, 320 + (1 - a2) * 30, "xbold", 88, WHITE, "mt", a2)
        p = ease_out((t - 1.7) / 1.9)
        if t > 1.7:
            n = int(round(525 * p))
            pulse = 1 + 0.08 * math.sin(clamp((t - 3.6) / 0.35) * math.pi)
            put(c, glow_sprite(360, RED, 70), cx, 650, "mm", alpha=p)
            txt(c, str(n), cx, 650, "black", 300, RED, "mm", 1.0, pulse)
        a3 = ap(t, 3.9)
        rich(c, [("bilinen CVE  ·  ", SUB, "semi"), ("14 uygulamada", PEACH, "semi"),
                 ("  ·  tek taramada", SUB, "semi")], cx, 850 + (1 - a3) * 20, 46, "mt", a3)

    scenes.append(Scene(5.8, hook, [(1.7, "riser"), (3.6, "pop")], "hook", "Giriş"))

    # 2 — Problem --------------------------------------------------------------
    apps = [("Google Chrome", "148.0.7778.217", "KRIT", "1 kritik · 1 yüksek · 6 orta"),
            ("AnyDesk", "9.6.1", "KRIT", "3 kritik · 6 yüksek · 1 orta"),
            ("Microsoft OneDrive", "26.084.0504", "KRIT", "1 kritik · 6 yüksek · 2 orta"),
            ("Node.js", "22.17.0", "KRIT", "1 kritik · 3 yüksek · 1 orta"),
            ("Adobe Acrobat Reader", "20.012.20043", "YÜKS", "247 yüksek · 135 orta"),
            ("Office 2019", "16.0.20026", "YÜKS", "14 yüksek · 2 orta"),
            ("VS Code", "1.123.0", "YÜKS", "12 yüksek · 2 orta"),
            ("Docker Desktop", "4.42.1", "YÜKS", "1 yüksek")]

    def problem(c, t):
        a = ap(t, 0.1)
        rich(c, [("Eski sürüm ", WHITE), ("=", OVER0), (" bilinen açık", RED)],
             cx, 110 + (1 - a) * -20, 76, "mt", a, "xbold")
        cw_, ch_, gap = 400, 180, 30
        x0 = (W - (4 * cw_ + 3 * gap)) / 2
        for i, (name, ver, sev, cnt) in enumerate(apps):
            e = ap(t, 0.6 + i * 0.13, 0.5)
            x = x0 + (i % 4) * (cw_ + gap)
            y = 280 + (i // 4) * (ch_ + gap) + (1 - e) * 40
            card(c, x, y, cw_, ch_, 20, MANTLE, SURF0, e)
            txt(c, name, x + 26, y + 24, "semi", 34, WHITE, alpha=e)
            txt(c, ver, x + 26, y + 72, "mono", 26, OVER0, alpha=e)
            pw, _ = pill(c, sev, x + 26, y + 124, 22, CRUST, SEV_COLOR[sev], None, "lt", e, "bold",
                         padx=12, pady=4)
            txt(c, cnt, x + 26 + pw + 14, y + 126, "med", 23, SUB, alpha=e)
        a4, a5 = ap(t, 3.7), ap(t, 4.4)
        txt(c, "Hepsini elle takip etmek mi?", cx, 760 + (1 - a4) * 25, "bold", 58, WHITE, "mt", a4)
        txt(c, "Kimse istemez.", cx, 845 + (1 - a5) * 25, "xbold", 58, PEACH, "mt", a5)

    scenes.append(Scene(7.2, problem, [(0.6 + i * 0.13, "tick") for i in range(8)], "problem",
                        "Sorun: eski sürümler"))

    # 3 — Intro ----------------------------------------------------------------
    def intro(c, t):
        e = ease_back(clamp((t - 0.15) / 0.6))
        put(c, glow_sprite(260, DUCK, 60), cx, 300, "mm", alpha=clamp(e))
        put(c, emoji_sprite("🦆", 220), cx, 300, "mm", scale=max(e, 0.01))
        a = ap(t, 0.7)
        txt(c, "mr-ducksafe", cx, 520 + (1 - a) * 30, "monob", 124, DUCK, "mm", a)
        a2 = ap(t, 1.2)
        txt(c, "Windows'taki kurulu yazılımlar için NVD CVE tarayıcı", cx, 640 + (1 - a2) * 20,
            "med", 46, SUB, "mt", a2)
        labels = ["Tek .exe", "Kurulum yok", "Hesap yok", "Açık kaynak · MIT"]
        widths = [tw(s, "semi", 34) + 52 for s in labels]
        x = cx - (sum(widths) + 24 * 3) / 2
        for i, s in enumerate(labels):
            ea = ap(t, 1.8 + i * 0.18)
            pill(c, s, x, 770 + (1 - ea) * 20, 34, TEXT, SURF0, SURF1, "lt", ea, "semi", padx=26)
            x += widths[i] + 24

    scenes.append(Scene(5.8, intro, [(0.15, "pop")] + [(1.8 + i * 0.18, "tick") for i in range(4)],
                        "intro", "mr-ducksafe nedir?"))

    # 4 — How it works -----------------------------------------------------------
    steps = [("🔍", "Tara", ["Windows Registry'den", "kurulu yazılımları okur"], "143 program"),
             ("🔗", "Eşleştir", ["50 hazır eşleme ile", "uygulamayı CPE'ye bağlar"], "21 eşleşme"),
             ("🌐", "Sorgula", ["NVD CVE API 2.0'da", "sürüm aralığına bakar"], "525 CVE"),
             ("🧠", "Açıkla", ["Riski sade dille anlatır,", "çözüm komutunu verir"], "14 aksiyon")]

    def how(c, t):
        a = ap(t, 0.1)
        txt(c, "Nasıl çalışır?", cx, 110 + (1 - a) * -20, "xbold", 76, WHITE, "mt", a)
        cw_, ch_, gap = 380, 430, 60
        x0 = (W - (4 * cw_ + 3 * gap)) / 2
        d = ImageDraw.Draw(c)
        for i, (em, title, desc, stat) in enumerate(steps):
            st = 0.5 + i * 1.25
            e = ap(t, st, 0.5)
            x, y = x0 + i * (cw_ + gap), 290 + (1 - e) * 40
            card(c, x, y, cw_, ch_, 24, MANTLE, SURF0, e)
            pill(c, str(i + 1), x + 28, y + 28, 26, CRUST, DUCK, None, "lt", e, "bold", padx=14, pady=4)
            put(c, emoji_sprite(em, 92), x + cw_ / 2, y + 120, "mm", e)
            txt(c, title, x + cw_ / 2, y + 190, "bold", 48, WHITE, "mt", e)
            for j, ln in enumerate(desc):
                txt(c, ln, x + cw_ / 2, y + 262 + j * 40, "reg", 29, SUB, "mt", e)
            txt(c, stat, x + cw_ / 2, y + 362, "monob", 30, TEAL, "mt", e)
            if i < 3:
                ea = clamp((t - st - 0.4) / 0.5)
                if ea > 0:
                    ax0 = x + cw_ + 10
                    ax1 = ax0 + (gap - 20) * ease_out(ea)
                    ay = 290 + ch_ / 2
                    d.line([ax0, ay, ax1, ay], fill=mix(BASE, DUCK, ea), width=5)
                    if ea >= 1:
                        d.polygon([(ax1 + 4, ay), (ax1 - 10, ay - 10), (ax1 - 10, ay + 10)], fill=DUCK)
        a2 = ap(t, 6.3)
        rich(c, [("Tarama tamamen yerel — NVD'ye yalnızca ", SUB, "med"),
                 ("{vendor, product, version}", TEAL, "mono"), (" gider.", SUB, "med")],
             cx, 790 + (1 - a2) * 20, 36, "mt", a2)

    scenes.append(Scene(9.2, how, [(0.5 + i * 1.25, "tick") for i in range(4)], "how", "Nasıl çalışır?"))

    # 5 — Terminal: scan -------------------------------------------------------------
    summary = load_capture("01-scan.txt")
    sc = TermScript()
    sc.wait(0.3)
    sc.type("mr-ducksafe scan", cps=16, pre=0.7)
    sc.out(BANNER)
    sc.spin("Kurulu yazılımlar taranıyor...", 1.0, done("Tarama tamamlandı — 143 program"))
    sc.spin("Eşleştiriliyor...", 0.6, done("Eşleştirme tamamlandı — 21 eşleşme"))
    sc.count("NVD sorgusu... {}/21", 21, 2.0, done("Tamamlandı — 525 bulgu"))
    sc.out([[S("")]])
    sc.out(summary[:21], rate=7)
    sc.out(summary[21:], rate=26)
    t_done = sc.wait(4.2) - 4.2

    def scan(c, t):
        caption(c, W, t, [("Tek komut: ", WHITE, "bold"), ("mr-ducksafe scan", DUCK, "monob")])
        draw_terminal(c, (150, 140, W - 150, H - 40), sc, t, 21)
        e = ease_back(clamp((t - t_done - 0.2) / 0.55))
        if e > 0:
            x, y, w, h = W - 150 - 40 - 600, 200, 600, 230
            a = clamp(e)
            card(c, x, y, w, h, 26, (32, 26, 40), RED, a)
            rich(c, [("525 açık", RED, "black")], x + w / 2, y + 30, 84, "mt", a)
            rich(c, [("→  ", OVER0, "bold"), ("14 aksiyon", GREEN, "black")], x + w / 2, y + 132,
                 60, "mt", a)

    scenes.append(Scene(sc.duration, scan, [(x, "click") for x in sc.clicks()]
                        + [(t_done + 0.2, "pop")], "scan", "Demo: mr-ducksafe scan"))

    # 6 — Plain language ------------------------------------------------------------------
    def plain(c, t):
        a = ap(t, 0.1)
        rich(c, [("Jargon yok. ", WHITE), ("Sade dil + çözüm komutu.", GREEN)],
             cx, 110 + (1 - a) * -20, 70, "mt", a, "xbold")
        e1 = ap(t, 0.5, 0.5)
        x, y, w, h = 110, 280 + (1 - e1) * 30, 760, 520
        card(c, x, y, w, h, 24, MANTLE, SURF0, e1)
        txt(c, "NVD kaydı", x + 40, y + 36, "semi", 30, OVER0, alpha=e1)
        txt(c, "CVE-2013-6647", x + 40, y + 90, "monob", 40, SKY, alpha=e1)
        for j, ln in enumerate(["A use-after-free in", "AnimationController::",
                                "endAnimationUpdate in Goog…"]):
            txt(c, ln, x + 40, y + 170 + j * 58, "mono", 34, OVER0, alpha=e1)
        txt(c, "Teknik, uzun, belirsiz.", x + 40, y + 430, "semi", 32, SUB, alpha=e1)
        ea = ap(t, 1.4, 0.4)
        put(c, text_sprite("→", "black", 110, DUCK), W / 2 + 0, 540, "mm", ea, 0.6 + 0.4 * ea)
        e2 = ap(t, 1.7, 0.5)
        x, y, w = W - 110 - 760, 280 + (1 - e2) * 30, 760
        card(c, x, y, w, h, 24, (26, 34, 36), GREEN, e2)
        txt(c, "mr-ducksafe", x + 40, y + 36, "semi", 30, DUCK, alpha=e2)
        pw, _ = pill(c, "KRIT", x + 40, y + 92, 28, CRUST, RED, None, "lt", e2, "bold", 14, 4)
        txt(c, "Google Chrome", x + 56 + pw, y + 90, "bold", 38, WHITE, alpha=e2)
        txt(c, "Acilen güncelle — aktif sömürülebilir.", x + 40, y + 170, "semi", 32, RED, alpha=e2)
        txt(c, "Bellek hatası yoluyla uygulama", x + 40, y + 232, "med", 34, TEXT, alpha=e2)
        txt(c, "ele geçirilebilir.", x + 40, y + 278, "med", 34, TEXT, alpha=e2)
        e3 = ap(t, 2.6, 0.5)
        put(c, panel_sprite(680, 86, 14, CRUST, SURF0), x + 40, y + 370, alpha=e3)
        rich(c, [("> ", OVER0, "monob"), ("winget upgrade --id Google.Chrome", TEAL, "monob")],
             x + 66, y + 392, 30, "lt", e3)

    scenes.append(Scene(7.0, plain, [(0.5, "tick"), (1.7, "tick"), (2.6, "pop")], "plain",
                        "Sade dilde risk + çözüm"))

    # 7 — Terminal: app details ---------------------------------------------------------------
    table = load_capture("02-chrome.txt")
    sd = TermScript()
    sd.wait(0.2)
    sd.type('mr-ducksafe scan --app "Chrome" --details', cps=28, pre=0.4)
    sd.out(BANNER)
    sd.spin("Kurulu yazılımlar taranıyor...", 0.35, done("Tarama tamamlandı — 143 program"))
    sd.spin("Eşleştiriliyor...", 0.3, done("Eşleştirme tamamlandı — 21 eşleşme"))
    sd.count("NVD sorgusu... {}/21", 21, 0.6, done("Tamamlandı — 525 bulgu"))
    sd.out([[S("")]])
    sd.out(table, rate=20)
    sd.wait(3.0)

    def details(c, t):
        caption(c, W, t, [("Tek uygulama, tüm CVE'ler: ", WHITE, "bold"), ("--app", DUCK, "monob"),
                          (" + ", OVER0, "bold"), ("--details", DUCK, "monob")])
        draw_terminal(c, (110, 140, W - 110, H - 40), sd, t, 17)

    scenes.append(Scene(sd.duration, details, [(x, "click") for x in sd.clicks()], "details",
                        "Demo: --app ve --details"))

    # 8 — CI ---------------------------------------------------------------------------------
    ci = TermScript()
    ci.wait(0.2)
    ci.type("mr-ducksafe scan --fail-on=high --format=json -o report.json", cps=30, pre=0.4)
    ci.spin("Kurulu yazılımlar taranıyor...", 0.4, done("Tarama tamamlandı — 143 program"))
    ci.spin("Eşleştiriliyor...", 0.3, done("Eşleştirme tamamlandı — 21 eşleşme"))
    ci.count("NVD sorgusu... {}/21", 21, 0.8, done("Tamamlandı — 525 bulgu"))
    ci.out([[S("Rapor yazıldı: report.json", SUB)]])
    ci.type("$LASTEXITCODE", cps=18, pre=0.4)
    ci.out([[S("1", RED, "b")], [S("")], [S("# yüksek+ bulgu var → exit 1 → pipeline 'failed'", OVER0)]])
    ci.wait(3.5)
    yaml = [
        "name: cve-scan",
        "on:",
        "  schedule: [{ cron: \"0 6 * * 1\" }]",
        "  workflow_dispatch:",
        "jobs:",
        "  scan:",
        "    runs-on: windows-latest",
        "    steps:",
        "      - uses: actions/setup-go@v5",
        "        with: { go-version: \"1.25\" }",
        "      - run: go install github.com/yemresalcan/mr-ducksafe/cmd/mr-ducksafe@latest",
        "      - run: mr-ducksafe scan --fail-on=high --format=json -o report.json",
        "        env:",
        "          NVD_API_KEY: ${{ secrets.NVD_API_KEY }}",
        "      - uses: actions/upload-artifact@v4",
        "        if: always()",
        "        with: { name: cve-report, path: report.json }",
    ]

    def yaml_segs(ln):
        m = re.match(r"^(\s*)(- )?([\w-]+):(.*)$", ln)
        if not m:
            return [S(ln, TEXT)]
        lead, dash, key, rest = m.groups()
        segs = [S(lead), S(dash or "", OVER0), S(key, MAUVE), S(":", OVER0)]
        for tok in re.split(r"(\"[^\"]*\"|\$\{\{[^}]*\}\})", rest):
            if tok.startswith('"'):
                segs.append(S(tok, GREEN))
            elif tok.startswith("${{"):
                segs.append(S(tok, PEACH))
            else:
                segs.append(S(tok, TEXT))
        return segs

    yaml_lines = [yaml_segs(ln) for ln in yaml]

    def cistep(c, t):
        caption(c, W, t, [("CI/CD için: ", WHITE, "bold"), ("--fail-on", DUCK, "monob"),
                          (" + JSON rapor", WHITE, "bold")])
        draw_terminal(c, (90, 150, 920, 720), ci, t, 18)
        e = ap(t, 0.8, 0.6)
        n = int(clamp((t - 0.8) * 9, 0, len(yaml_lines)))
        draw_terminal(c, (950, 150 + (1 - e) * 30, W - 90, 720 + (1 - e) * 30), None, t, 16,
                      ".github/workflows/cve-scan.yml", yaml_lines[:n], alpha=e)
        codes = [("0", GREEN, "Temiz", "eşik aşılmadı"), ("1", RED, "Eşik aşıldı", "--fail-on tetiklendi"),
                 ("2", PEACH, "Tarama hatası", "NVD'ye ulaşılamadı vb.")]
        cw_ = (W - 180 - 2 * 30) / 3
        for i, (code, col, title, desc) in enumerate(codes):
            ea = ap(t, 2.0 + i * 0.25)
            x, y = 90 + i * (cw_ + 30), 770 + (1 - ea) * 25
            card(c, x, y, cw_, 200, 22, MANTLE, SURF0, ea)
            txt(c, "exit", x + 36, y + 44, "mono", 30, OVER0, alpha=ea)
            txt(c, code, x + 36, y + 80, "black", 96, col, alpha=ea)
            txt(c, title, x + 150, y + 62, "bold", 42, WHITE, alpha=ea)
            txt(c, desc, x + 150, y + 120, "med", 30, SUB, alpha=ea)

    scenes.append(Scene(ci.duration, cistep, [(x, "click") for x in ci.clicks()]
                        + [(2.0 + i * 0.25, "tick") for i in range(3)], "ci",
                        "CI/CD: --fail-on ve JSON"))

    # 9 — History + privacy ----------------------------------------------------------------------
    hist = load_capture("03-history.txt")
    hs = TermScript()
    hs.wait(0.2)
    hs.type("mr-ducksafe history --limit 5", cps=24, pre=0.4)
    hs.out(hist, rate=10)
    hs.wait(0.5)
    hs.type("mr-ducksafe prune --older-than=30d", cps=24, pre=0.6)
    hs.wait(4.0)
    privacy = ["Tarama tamamen yerel", "NVD'ye yalnızca {vendor, product, version}",
               "Makine adı, kullanıcı, IP gönderilmez", "Telemetri yok",
               "Geçmiş yerel SQLite'ta: ~/.mr-ducksafe"]

    def histpriv(c, t):
        caption(c, W, t, [("Yerel geçmiş ", WHITE, "bold"), ("·", OVER0, "bold"),
                          (" önce gizlilik", GREEN, "bold")])
        draw_terminal(c, (90, 160, 960, 620), hs, t, 21)
        x = 1040
        a = ap(t, 0.6)
        put(c, emoji_sprite("🔒", 60), x, 200, "lt", a)
        txt(c, "Gizlilik", x + 80, 196, "xbold", 60, WHITE, alpha=a)
        for i, s in enumerate(privacy):
            e = ap(t, 1.0 + i * 0.35)
            y = 320 + i * 92 + (1 - e) * 20
            put(c, panel_sprite(46, 46, 23, (36, 52, 44), GREEN), x, y, alpha=e)
            txt(c, "✓", x + 23, y + 23, "monob", 30, GREEN, "mm", e)
            txt(c, s, x + 68, y + 2, "semi", 34, TEXT, alpha=e)
        e = ap(t, 2.0)
        card(c, 90, 680 + (1 - e) * 20, 870, 300, 22, MANTLE, SURF0, e)
        txt(c, "Her tarama SQLite'a kaydedilir:", 130, 715 + (1 - e) * 20, "semi", 32, WHITE, alpha=e)
        for j, (cmd, desc) in enumerate((("history", "geçmiş taramalar"),
                                         ("prune --older-than=30d", "eski kayıtları sil"),
                                         ("--db PATH", "farklı DB yolu"))):
            yy = 785 + j * 60 + (1 - e) * 20
            txt(c, cmd, 130, yy, "monob", 28, DUCK, alpha=e)
            txt(c, "— " + desc, 130 + tw(cmd, "monob", 28) + 16, yy + 2, "med", 28, SUB, alpha=e)

    scenes.append(Scene(hs.duration, histpriv, [(x, "click") for x in hs.clicks()], "history",
                        "Geçmiş ve gizlilik"))

    # 10 — Outro --------------------------------------------------------------------------------------
    def outro(c, t):
        e = ease_back(clamp((t - 0.1) / 0.6))
        put(c, glow_sprite(200, DUCK, 50), cx, 180, "mm", alpha=clamp(e))
        put(c, emoji_sprite("🦆", 150), cx, 180, "mm", scale=max(e, 0.01))
        a = ap(t, 0.4)
        txt(c, "mr-ducksafe", cx, 340 + (1 - a) * 20, "monob", 104, DUCK, "mm", a)
        a = ap(t, 0.8)
        txt(c, "Bilinen açıkları bulmak ve kapatmak hiç bu kadar basit olmamıştı.", cx,
            420 + (1 - a) * 20, "med", 40, SUB, "mt", a)
        e2 = ap(t, 1.3)
        x, y, w, h = 300, 520 + (1 - e2) * 30, W - 600, 270
        card(c, x, y, w, h, 22, CRUST, SURF0, e2)
        rows = [[("# Windows: GitHub Releases → mr-ducksafe.exe indir", OVER0, "mono")],
                [("go install ", TEXT, "mono"),
                 ("github.com/yemresalcan/mr-ducksafe/cmd/mr-ducksafe@latest", TEAL, "mono")],
                [("mr-ducksafe scan", DUCK, "monob")]]
        for j, r in enumerate(rows):
            rich(c, r, x + 50, y + 40 + j * 72, 30, "lt", e2)
        e3 = ap(t, 2.2)
        pill(c, "github.com/yemresalcan/mr-ducksafe", cx, 850 + (1 - e3) * 20, 48, CRUST, DUCK,
             None, "mt", e3, "xbold")
        e4 = ap(t, 2.8)
        s4 = "Yıldız verin  ·  Yeni mapping PR'larına açık  ·  MIT"
        w4 = tw(s4, "semi", 36) + 60
        put(c, emoji_sprite("⭐", 44), cx - w4 / 2, 985, "lm", e4)
        txt(c, s4, cx - w4 / 2 + 60, 985, "semi", 36, TEXT, "lm", e4)

    scenes.append(Scene(8.5, outro, [(0.1, "pop"), (2.2, "tick")], "outro", "Kurulum ve bağlantılar"))
    return scenes


# ======================= Shorts (9:16) scenes ============================
def short_scenes(W=1080, H=1920):
    scenes = []
    cx = W / 2

    def hook(c, t):
        a1, a2 = ap(t, 0.15), ap(t, 0.5)
        txt(c, "Bilgisayarında", cx, 430 + (1 - a1) * 30, "bold", 72, SUB, "mt", a1)
        txt(c, "kaç güvenlik", cx, 530 + (1 - a2) * 30, "black", 104, WHITE, "mt", a2)
        txt(c, "açığı var?", cx, 650 + (1 - a2) * 30, "black", 104, WHITE, "mt", a2)
        p = ease_out((t - 1.3) / 1.7)
        if t > 1.3:
            pulse = 1 + 0.08 * math.sin(clamp((t - 3.0) / 0.35) * math.pi)
            put(c, glow_sprite(320, RED, 70), cx, 1000, "mm", alpha=p)
            txt(c, str(int(round(525 * p))), cx, 1000, "black", 300, RED, "mm", 1, pulse)
        a3 = ap(t, 3.2)
        txt(c, "bilinen CVE · tek taramada", cx, 1190 + (1 - a3) * 20, "semi", 50, PEACH, "mt", a3)

    scenes.append(Scene(4.8, hook, [(1.3, "riser"), (3.0, "pop")]))

    summary = load_capture("01-scan.txt")
    sc = TermScript()
    sc.wait(0.2)
    sc.type("mr-ducksafe scan", cps=18, pre=0.4)
    sc.out(BANNER)
    sc.spin("Kurulu yazılımlar taranıyor...", 0.6, done("Tarama tamamlandı — 143 program"))
    sc.spin("Eşleştiriliyor...", 0.4, done("Eşleştirme tamamlandı — 21 eşleşme"))
    sc.count("NVD sorgusu... {}/21", 21, 1.2, done("Tamamlandı — 525 bulgu"))
    sc.out([[S("")]])
    sc.out(summary[:27], rate=10)
    sc.out(summary[27:], rate=30)
    t_done = sc.wait(3.6) - 3.6

    def scan(c, t):
        a = ap(t, 0.1)
        txt(c, "Tek komut:", cx, 190 + (1 - a) * -20, "bold", 56, WHITE, "mt", a)
        pill(c, "mr-ducksafe scan", cx, 272 + (1 - a) * -20, 50, CRUST, DUCK, None, "mt", a, "monob")
        draw_terminal(c, (20, 400, W - 20, 1560), sc, t, 17)
        e = ease_back(clamp((t - t_done - 0.2) / 0.55))
        if e > 0:
            x, y, w, h = 90, 1180, W - 180, 300
            card(c, x, y, w, h, 28, (32, 26, 40), RED, clamp(e))
            txt(c, "525 açık", cx, y + 40, "black", 110, RED, "mt", clamp(e))
            rich(c, [("→  ", OVER0, "bold"), ("14 aksiyon", GREEN, "black")], cx, y + 180, 72,
                 "mt", clamp(e))

    scenes.append(Scene(sc.duration, scan, [(x, "click") for x in sc.clicks()] + [(t_done + 0.2, "pop")]))

    def plain(c, t):
        a = ap(t, 0.1)
        txt(c, "Jargon yok.", cx, 300 + (1 - a) * -20, "black", 96, WHITE, "mt", a)
        txt(c, "Sade dil + çözüm.", cx, 410 + (1 - a) * -20, "black", 76, GREEN, "mt", a)
        e1 = ap(t, 0.4)
        x, y, w = 70, 560 + (1 - e1) * 30, W - 140
        card(c, x, y, w, 300, 24, MANTLE, SURF0, e1)
        txt(c, "NVD kaydı · CVE-2013-6647", x + 40, y + 34, "semi", 32, OVER0, alpha=e1)
        for j, ln in enumerate(["A use-after-free in", "AnimationController::endAnim…"]):
            txt(c, ln, x + 40, y + 110 + j * 60, "mono", 36, OVER0, alpha=e1)
        ea = ap(t, 1.0)
        txt(c, "↓", cx, 935, "black", 90, DUCK, "mm", ea)
        e2 = ap(t, 1.3)
        y = 1010 + (1 - e2) * 30
        card(c, x, y, w, 470, 24, (26, 34, 36), GREEN, e2)
        pw, _ = pill(c, "KRIT", x + 40, y + 40, 32, CRUST, RED, None, "lt", e2, "bold", 14, 4)
        txt(c, "Google Chrome", x + 58 + pw, y + 36, "bold", 44, WHITE, alpha=e2)
        txt(c, "Bellek hatası yoluyla", x + 40, y + 130, "semi", 46, TEXT, alpha=e2)
        txt(c, "uygulama ele geçirilebilir.", x + 40, y + 190, "semi", 46, TEXT, alpha=e2)
        e3 = ap(t, 2.2)
        put(c, panel_sprite(w - 80, 110, 16, CRUST, SURF0), x + 40, y + 300, alpha=e3)
        txt(c, "winget upgrade --id Google.Chrome", x + 70, y + 333, "monob", 34, TEAL, alpha=e3)

    scenes.append(Scene(5.6, plain, [(0.4, "tick"), (1.3, "tick"), (2.2, "pop")]))

    feats = [("📦", "Tek .exe — kurulum yok"), ("🏛", "Resmi NVD CVE verisi"),
             ("🧠", "Riskler sade dilde"), ("🛠", "winget ile tek komut çözüm"),
             ("⚙", "CI için --fail-on"), ("🔒", "%100 yerel, telemetri yok")]

    def features(c, t):
        a = ap(t, 0.1)
        txt(c, "Neden", cx, 300 + (1 - a) * -20, "bold", 64, SUB, "mt", a)
        txt(c, "mr-ducksafe?", cx, 380 + (1 - a) * -20, "monob", 96, DUCK, "mt", a)
        for i, (em, s) in enumerate(feats):
            e = ap(t, 0.5 + i * 0.32)
            y = 560 + i * 150 + (1 - e) * 30
            card(c, 70, y, W - 140, 124, 24, MANTLE, SURF0, e, shadow=False)
            put(c, emoji_sprite(em, 64), 120, y + 62, "lm", e)
            txt(c, s, 220, y + 62, "semi", 46, TEXT, "lm", e)

    scenes.append(Scene(6.4, features, [(0.5 + i * 0.32, "tick") for i in range(6)]))

    def cta(c, t):
        e = ease_back(clamp((t - 0.1) / 0.6))
        put(c, glow_sprite(280, DUCK, 60), cx, 560, "mm", alpha=clamp(e))
        put(c, emoji_sprite("🦆", 260), cx, 560, "mm", scale=max(e, 0.01))
        a = ap(t, 0.5)
        txt(c, "mr-ducksafe", cx, 820 + (1 - a) * 20, "monob", 112, DUCK, "mm", a)
        a = ap(t, 0.9)
        txt(c, "Ücretsiz · Açık kaynak · MIT", cx, 920 + (1 - a) * 20, "semi", 50, TEXT, "mt", a)
        a = ap(t, 1.4)
        pill(c, "github.com/yemresalcan/mr-ducksafe", cx, 1040 + (1 - a) * 20, 44, CRUST, DUCK,
             None, "mt", a, "xbold", padx=34, pady=18)
        a = ap(t, 2.0)
        s5 = "Yıldızla · kaydet · paylaş"
        w5 = tw(s5, "bold", 52) + 72
        put(c, emoji_sprite("⭐", 52), cx - w5 / 2, 1220, "lm", a)
        txt(c, s5, cx - w5 / 2 + 72, 1220, "bold", 52, WHITE, "lm", a)

    scenes.append(Scene(6.0, cta, [(0.1, "pop"), (1.4, "tick")]))
    return scenes


# --------------------------------------------------------------------------
# Audio
# --------------------------------------------------------------------------
def mtof(m):
    return 440.0 * 2 ** ((m - 69) / 12)


def synth_music(total, bpm=100, seed=7):
    rng = np.random.default_rng(seed)
    n = int(total * SR) + SR
    L, R = np.zeros(n), np.zeros(n)
    beat = 60 / bpm
    bar = 4 * beat
    chords = [[57, 60, 64, 67], [53, 57, 60, 64], [48, 52, 55, 59], [55, 59, 62, 66]]  # Am7 Fmaj7 Cmaj7 G
    roots = [45, 41, 48, 43]
    nbars = int(math.ceil(total / bar))

    def add(sig, start, l=1.0, r=1.0):
        i = int(start * SR)
        if i >= n:
            return
        sig = sig[: n - i]
        L[i:i + len(sig)] += sig * l
        R[i:i + len(sig)] += sig * r

    for b in range(nbars):
        s = b * bar
        ch = chords[b % 4]
        # pad
        dur = bar + 0.6
        tt = np.arange(int(dur * SR)) / SR
        env = np.clip(tt / 0.5, 0, 1) * np.clip((dur - tt) / 0.7, 0, 1)
        for m in ch:
            f = mtof(m)
            for det, pan in ((0.997, 0.8), (1.003, 1.2)):
                sig = np.sin(2 * np.pi * f * det * tt) + 0.25 * np.sin(4 * np.pi * f * det * tt)
                add(sig * env * 0.028, s, 2 - pan, pan)
        drums = 0 < b < nbars - 1
        # bass
        for k in range(8):
            if k % 2 and not drums:
                continue
            bt = np.arange(int(beat * 0.5 * SR)) / SR
            f = mtof(roots[b % 4])
            sig = (np.sin(2 * np.pi * f * bt) + 0.3 * np.sin(4 * np.pi * f * bt)) * np.exp(-bt * 6)
            add(sig * (0.13 if k % 2 == 0 else 0.07), s + k * beat / 2)
        # arpeggio
        for k in range(8):
            m = ch[[0, 1, 2, 3, 2, 1, 3, 2][k]] + 12
            at = np.arange(int(0.4 * SR)) / SR
            sig = np.sin(2 * np.pi * mtof(m) * at) * np.exp(-at * 8) * np.clip(at / 0.005, 0, 1)
            pan = 0.6 if k % 2 else 1.4
            add(sig * 0.035, s + k * beat / 2, 2 - pan, pan)
        if drums:
            for k in range(4):
                if k in (0, 2):
                    kt = np.arange(int(0.35 * SR)) / SR
                    ph = 2 * np.pi * np.cumsum(48 + 90 * np.exp(-kt * 30)) / SR
                    add(np.sin(ph) * np.exp(-kt * 9) * 0.32, s + k * beat)
                if k in (1, 3):
                    st = np.arange(int(0.2 * SR)) / SR
                    nz = np.diff(rng.standard_normal(len(st) + 1))
                    add(nz * np.exp(-st * 22) * 0.03, s + k * beat)
                ht = np.arange(int(0.06 * SR)) / SR
                hz = np.diff(np.diff(rng.standard_normal(len(ht) + 2)))
                add(hz * np.exp(-ht * 70) * 0.012, s + k * beat + beat / 2, 0.8, 1.2)
    fade = np.clip((total - np.arange(n) / SR) / 2.5, 0, 1) * np.clip(np.arange(n) / SR / 0.8, 0, 1)
    return L * fade, R * fade


def sfx_layer(events, total, seed=3):
    rng = np.random.default_rng(seed)
    n = int(total * SR) + SR
    out = np.zeros(n)

    def add(sig, start):
        i = int(start * SR)
        if 0 <= i < n:
            sig = sig[: n - i]
            out[i:i + len(sig)] += sig

    for t, kind in events:
        if kind == "click":
            ct = np.arange(int(0.025 * SR)) / SR
            nz = np.diff(rng.standard_normal(len(ct) + 1))
            body = np.sin(2 * np.pi * rng.uniform(1600, 2400) * ct)
            add((nz * 0.5 + body * 0.5) * np.exp(-ct * 260) * rng.uniform(0.10, 0.16), t)
        elif kind == "tick":
            ct = np.arange(int(0.08 * SR)) / SR
            add(np.sin(2 * np.pi * 1320 * ct) * np.exp(-ct * 55) * 0.05, t)
        elif kind == "pop":
            ct = np.arange(int(0.25 * SR)) / SR
            f = 520 + 500 * (1 - np.exp(-ct * 30))
            add(np.sin(2 * np.pi * np.cumsum(f) / SR) * np.exp(-ct * 14) * 0.12, t)
        elif kind == "whoosh":
            d = 0.55
            ct = np.arange(int(d * SR)) / SR
            nz = rng.standard_normal(len(ct))
            k = 40
            nz = np.convolve(nz, np.ones(k) / k, mode="same") * 4
            env = np.sin(np.pi * ct / d) ** 2
            add(nz * env * 0.06, t - d / 2)
        elif kind == "riser":
            d = 1.9
            ct = np.arange(int(d * SR)) / SR
            f = 180 + 500 * (ct / d) ** 2
            sig = np.sin(2 * np.pi * np.cumsum(f) / SR) * (ct / d) ** 2 * 0.05
            add(sig, t)
    return out


def write_wav(path, scenes, total):
    L, R = synth_music(total)
    events = []
    t0 = 0.0
    for i, sc in enumerate(scenes):
        if i:
            events.append((t0, "whoosh"))
        events += [(t0 + lt, k) for lt, k in sc.sfx]
        t0 += sc.dur
    fx = sfx_layer(events, total)
    n = min(len(L), len(fx))
    L, R = L[:n] + fx[:n], R[:n] + fx[:n]
    peak = max(np.abs(L).max(), np.abs(R).max(), 1e-6)
    gain = 0.85 / peak
    L, R = np.tanh(L * gain * 1.1), np.tanh(R * gain * 1.1)
    data = (np.stack([L, R], 1)[: int(total * SR)] * 32000).astype(np.int16)
    with wave.open(str(path), "wb") as w:
        w.setnchannels(2)
        w.setsampwidth(2)
        w.setframerate(SR)
        w.writeframes(data.tobytes())


# --------------------------------------------------------------------------
# Rendering
# --------------------------------------------------------------------------
_G = {}


def frame_at(t):
    W, H, scenes = _G["W"], _G["H"], _G["scenes"]
    bg = background(W, H)
    acc = 0.0
    for i, sc in enumerate(scenes):
        if t < acc + sc.dur or i == len(scenes) - 1:
            lt = min(t - acc, sc.dur)
            img = bg.copy()
            sc.fn(img, lt)
            fin = 0.6 if i == 0 else 0.25
            fout = 0.8 if i == len(scenes) - 1 else 0.25
            a = clamp(lt / fin) * clamp((sc.dur - lt) / fout)
            if a < 1:
                img = Image.blend(bg, img, a)
            return img
        acc += sc.dur


def _render_bytes(i):
    return frame_at(i / FPS).tobytes()


def render_video(kind):
    W, H, scenes = (1920, 1080, yt_scenes()) if kind == "youtube" else (1080, 1920, short_scenes())
    _G.update(W=W, H=H, scenes=scenes)
    total = sum(s.dur for s in scenes)
    nframes = int(round(total * FPS))
    OUT.mkdir(parents=True, exist_ok=True)
    out = OUT / f"mr-ducksafe-{kind}.mp4"
    print(f"[{kind}] {W}x{H}, {total:.1f}s, {nframes} frames")
    acc = 0.0
    for s in scenes:
        if s.chapter:
            print(f"   {int(acc // 60)}:{int(acc % 60):02d}  {s.chapter}")
        acc += s.dur
    with tempfile.TemporaryDirectory() as td:
        wav = Path(td) / "audio.wav"
        write_wav(wav, scenes, total)
        cmd = ["ffmpeg", "-y", "-loglevel", "error", "-f", "rawvideo", "-pix_fmt", "rgb24",
               "-s", f"{W}x{H}", "-r", str(FPS), "-i", "-", "-i", str(wav),
               "-c:v", "libx264", "-preset", "slow", "-crf", "21", "-pix_fmt", "yuv420p",
               "-profile:v", "high", "-c:a", "aac", "-b:a", "160k", "-shortest",
               "-movflags", "+faststart", str(out)]
        proc = subprocess.Popen(cmd, stdin=subprocess.PIPE)
        ctx = mp.get_context("fork")
        with ctx.Pool(max(1, os.cpu_count() - 1)) as pool:
            for k, buf in enumerate(pool.imap(_render_bytes, range(nframes), chunksize=6)):
                proc.stdin.write(buf)
                if k % 300 == 0:
                    print(f"   frame {k}/{nframes}", flush=True)
        proc.stdin.close()
        if proc.wait() != 0:
            raise SystemExit("ffmpeg failed")
    print(f"wrote {out} ({out.stat().st_size / 1e6:.1f} MB)")


def preview(kind, times=None):
    W, H, scenes = (1920, 1080, yt_scenes()) if kind == "youtube" else (1080, 1920, short_scenes())
    _G.update(W=W, H=H, scenes=scenes)
    pdir = Path(tempfile.gettempdir()) / "ducksafe-preview"
    pdir.mkdir(exist_ok=True)
    acc = 0.0
    for i, s in enumerate(scenes):
        for frac in (times or [0.85]):
            t = acc + s.dur * frac
            p = pdir / f"{kind}-{i:02d}-{int(frac * 100)}.png"
            frame_at(t).save(p)
            print(p)
        acc += s.dur


# --------------------------------------------------------------------------
# Thumbnail
# --------------------------------------------------------------------------
def render_thumbnail():
    W, H = 1280, 720
    img = background(W, H).copy()
    put(img, glow_sprite(420, RED, 80), 250, 300, "mm")
    put(img, glow_sprite(330, DUCK, 70), 1060, 470, "mm")

    # tilted terminal on the right
    term = Image.new("RGB", (720, 470), BASE)
    lines = [colorize(l) for l in (DEMO / "01-scan.txt").read_text(encoding="utf-8-sig").splitlines()[:21]]
    tmp_bg = Image.new("RGB", (760, 510), BASE)
    draw_terminal(tmp_bg, (20, 20, 740, 490), None, 0, 15, "mr-ducksafe scan", lines)
    term = tmp_bg.convert("RGBA").rotate(-7, resample=Image.BICUBIC, expand=True,
                                          fillcolor=(0, 0, 0, 0))
    # make the BASE margin transparent again
    arr = np.array(term)
    m = (np.abs(arr[..., :3].astype(int) - np.array(BASE)).sum(-1) < 6)
    arr[m, 3] = 0
    term = Image.fromarray(arr)
    put(img, term, 600, 60, alpha=0.92)

    # duck
    put(img, emoji_sprite("🦆", 330), 1085, 520, "mm")

    stroke = 7
    pill(img, "WINDOWS · ÜCRETSİZ CVE TARAYICI", 56, 48, 30, CRUST, DUCK, None, "lt", 1, "xbold",
         padx=22, pady=10)
    x = 50
    w1 = tw("525 ", "black", 196)
    txt(img, "525", x, 100, "black", 196, RED, stroke=stroke)
    txt(img, "AÇIK", x + w1 - 8, 100, "black", 196, WHITE, stroke=stroke)
    txt(img, "→ 14 AKSİYON", x + 4, 335, "black", 110, GREEN, stroke=stroke)
    txt(img, "tek komutla, sade dilde", x + 8, 480, "bold", 54, WHITE, stroke=5)
    pill(img, "> mr-ducksafe scan", 58, 590, 40, DUCK, CRUST, SURF1, "lt", 1, "monob",
         padx=24, pady=12)
    OUT.mkdir(parents=True, exist_ok=True)
    p = OUT / "thumbnail.png"
    img.save(p, optimize=True)
    print(f"wrote {p}")


if __name__ == "__main__":
    args = sys.argv[1:]
    if args[:1] == ["preview"]:
        preview(args[1] if len(args) > 1 else "youtube",
                [float(x) for x in args[2:]] or None)
        sys.exit(0)
    targets = args or ["thumbnail", "youtube", "short"]
    for tgt in targets:
        if tgt == "thumbnail":
            render_thumbnail()
        elif tgt in ("youtube", "short"):
            render_video(tgt)
        else:
            raise SystemExit(f"unknown target {tgt!r}")
