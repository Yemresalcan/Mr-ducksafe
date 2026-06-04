"""Render a typewriter-style terminal demo GIF for mr-ducksafe.

Reads pre-captured command outputs and produces docs/demo.gif.
No screen recording — frames are rendered with Pillow.
"""
from __future__ import annotations

import os
import sys
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw, ImageFont
import imageio.v2 as imageio

ROOT = Path(__file__).resolve().parent
PROJECT = ROOT.parent.parent
CAPS = ROOT
OUT = PROJECT / "docs" / "demo.gif"

WIDTH, HEIGHT = 1280, 700
FONT_SIZE = 14
LEFT_PAD, TOP_PAD = 14, 12
LINE_SPACING = 6

# Catppuccin Mocha-ish palette
BG       = (30, 30, 46)
FG       = (205, 214, 244)
PROMPT   = (137, 220, 235)   # cyan
CMD      = (245, 224, 220)   # cream
KRIT     = (243, 139, 168)   # red
YUKS     = (250, 179, 135)   # orange
ORTA     = (249, 226, 175)   # yellow
DUSK     = (147, 153, 178)   # gray
HEADER   = (203, 166, 247)   # purple
ACCENT   = (148, 226, 213)   # teal
DIM      = (108, 112, 134)
GREEN    = (166, 227, 161)
RED_HOT  = (243, 139, 168)

# Try a few monospace fonts in order of preference.
FONT_CANDIDATES = [
    r"C:\Windows\Fonts\CascadiaMono.ttf",
    r"C:\Windows\Fonts\CascadiaCode.ttf",
    r"C:\Windows\Fonts\consola.ttf",
    r"C:\Windows\Fonts\cour.ttf",
]
FONT_BOLD_CANDIDATES = [
    r"C:\Windows\Fonts\CascadiaMono-Bold.ttf",
    r"C:\Windows\Fonts\consolab.ttf",
    r"C:\Windows\Fonts\courbd.ttf",
]

def load_font(candidates, size):
    for p in candidates:
        if os.path.exists(p):
            return ImageFont.truetype(p, size)
    return ImageFont.load_default()

FONT = load_font(FONT_CANDIDATES, FONT_SIZE)
FONT_BOLD = load_font(FONT_BOLD_CANDIDATES, FONT_SIZE)

# Measure character cell.
bbox = FONT.getbbox("M")
CHAR_W = bbox[2] - bbox[0]
CHAR_H = (bbox[3] - bbox[1]) + LINE_SPACING
ROWS = (HEIGHT - 2 * TOP_PAD) // CHAR_H

PROMPT_STR = "PS F:\\mr-ducksafe> "

def line_color(text: str):
    s = text.lstrip()
    if s.startswith("KRIT"): return KRIT
    if s.startswith("YÜKS") or s.startswith("YUKS"): return YUKS
    if s.startswith("ORTA"): return ORTA
    if s.startswith("DÜŞK") or s.startswith("DUSK"): return DUSK
    if "Aciliyet:" in s: return DIM
    if "Risk:" in s: return DIM
    if "Çözüm:" in s or "Cozum:" in s: return ACCENT
    if s.startswith("mr-ducksafe —") or s.startswith("mr-ducksafe scan #"):
        return HEADER
    if "Bulgular:" in s: return HEADER
    if "Detay:" in s: return DIM
    return FG

# Buffer of fully-committed lines: list[(text, color)]
buffer: list[tuple[str, tuple[int,int,int]]] = []
frames: list[Image.Image] = []

def render_buffer(extra_line: tuple[str, tuple[int,int,int]] | None = None, cursor: bool = False):
    img = Image.new("RGB", (WIDTH, HEIGHT), BG)
    draw = ImageDraw.Draw(img)

    # Title bar
    draw.rectangle([(0, 0), (WIDTH, 26)], fill=(17, 17, 27))
    draw.ellipse([(10, 8), (20, 18)], fill=(243, 139, 168))
    draw.ellipse([(28, 8), (38, 18)], fill=(249, 226, 175))
    draw.ellipse([(46, 8), (56, 18)], fill=(166, 227, 161))
    draw.text((WIDTH//2 - 50, 6), "mr-ducksafe", fill=DIM, font=FONT)

    content_top = 32
    visible = list(buffer)
    if extra_line is not None:
        visible.append(extra_line)
    visible = visible[-(ROWS-2):]

    for i, (text, color) in enumerate(visible):
        y = content_top + i * CHAR_H
        draw.text((LEFT_PAD, y), text, fill=color, font=FONT)

    if cursor and visible:
        last_text = visible[-1][0]
        cx = LEFT_PAD + len(last_text) * CHAR_W
        cy = content_top + (len(visible) - 1) * CHAR_H
        draw.rectangle([(cx, cy + 2), (cx + CHAR_W - 1, cy + CHAR_H - 2)], fill=FG)
    return img

def snap(extra=None, cursor=False, repeat=1):
    img = render_buffer(extra, cursor)
    for _ in range(repeat):
        frames.append(img)

def commit_line(text: str, color=None):
    buffer.append((text, color or line_color(text)))

def type_command(cmd: str, *, step=2):
    """Render typewriter typing of cmd, then commit the full prompt+cmd."""
    snap(extra=(PROMPT_STR, PROMPT), cursor=True, repeat=4)
    for i in range(0, len(cmd) + 1, step):
        partial = PROMPT_STR + cmd[:i]
        snap(extra=(partial, CMD), cursor=True, repeat=1)
    # final committed line
    commit_line(PROMPT_STR + cmd, CMD)
    snap(repeat=2)

def reveal_file(path: Path, *, max_lines: int = 22, hold: int = 8):
    text = path.read_text(encoding="utf-8", errors="replace")
    lines = text.splitlines()
    # Drop trailing empty lines
    while lines and not lines[-1].strip():
        lines.pop()
    truncated = False
    if len(lines) > max_lines:
        lines = lines[:max_lines]
        truncated = True
    for ln in lines:
        commit_line(ln)
        snap(repeat=1)
    if truncated:
        commit_line("  ...", DIM)
        snap(repeat=2)
    snap(repeat=hold)

def spinner_pause(label: str, ticks: int = 6):
    glyphs = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
    for i in range(ticks):
        snap(extra=(f"  {glyphs[i % len(glyphs)]} {label}", DIM), cursor=False, repeat=1)

def blank_line():
    commit_line("", FG)
    snap(repeat=1)

# === SCENARIO 1: full scan (guided summary) ============================
type_command("mr-ducksafe scan")
spinner_pause("NVD'den CVE verisi çekiliyor...", ticks=10)
reveal_file(CAPS / "01-scan.txt", max_lines=22, hold=12)

# === SCENARIO 2: filter to Chrome ======================================
blank_line()
type_command('mr-ducksafe scan --app "Chrome"')
spinner_pause("Chrome taranıyor...", ticks=4)
reveal_file(CAPS / "02-chrome.txt", max_lines=18, hold=10)

# === SCENARIO 3: history ==============================================
blank_line()
type_command("mr-ducksafe history --limit 5")
reveal_file(CAPS / "03-history.txt", max_lines=12, hold=8)

# === SCENARIO 4: fail-on=high (CI use-case) ============================
blank_line()
type_command("mr-ducksafe scan --fail-on=high")
spinner_pause("Tarama + eşik kontrolü...", ticks=5)
reveal_file(CAPS / "04-failon.txt", max_lines=16, hold=4)
commit_line("[exit code: 1]  →  CI bu işi 'failed' olarak işaretler.", RED_HOT)
snap(repeat=20)

# === EXPORT ============================================================
print(f"frames: {len(frames)}, char cell: {CHAR_W}x{CHAR_H}, rows: {ROWS}")
np_frames = [np.array(f) for f in frames]
# duration: per-frame in seconds (v2 API). 0.07s/frame ≈ 14 fps.
imageio.mimsave(OUT, np_frames, duration=0.07, loop=0, subrectangles=True)
size = OUT.stat().st_size
print(f"wrote {OUT} ({size/1024:.0f} KB)")
