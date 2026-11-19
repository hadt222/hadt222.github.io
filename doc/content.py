
from pathlib import Path
from html import escape
from urllib.parse import quote

folder = Path(__file__).resolve().parent

files = sorted(
    (f for f in folder.glob("*.html")
     if f.name.lower() != "index.html"),
    key=lambda f: f.name.lower()
)

items = "\n".join(
    f'<a href="{quote(f.name)}">{escape(f.stem)}</a>'
    for f in files
)

html = f"""<!DOCTYPE html>
<html lang="vi">
<head>
<meta charset="UTF-8">
<title>Mục lục HTML</title>
<style>
body {{
    font-family: Arial, sans-serif;
    max-width: 900px;
    margin: 40px auto;
    padding: 20px;
    background: #f8f9fa;
}}
h1 {{ color: #333; }}
a {{
    display: block;
    padding: 12px 16px;
    margin: 6px 0;
    background: white;
    border-radius: 6px;
    text-decoration: none;
    color: #0066cc;
    box-shadow: 0 1px 3px #ddd;
}}
a:hover {{
    background: #e6f0ff;
}}
</style>
</head>
<body>
<h1>📚 Mục lục HTML</h1>
<p>Tổng số trang: {len(files)}</p>
{items}
</body>
</html>"""

(folder / "index.html").write_text(
    html, encoding="utf-8"
)

print(f"Đã tạo index.html với {len(files)} trang")
