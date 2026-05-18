import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const docsDir = path.dirname(__dirname);

const colors = {
  blue: "#dbeafe",
  green: "#dcfce7",
  yellow: "#fef3c7",
  orange: "#ffedd5",
  red: "#fee2e2",
  purple: "#ede9fe",
  gray: "#f3f4f6",
  white: "#ffffff",
};

let seed = 100;
let index = 0;

function nextId(prefix = "el") {
  index += 1;
  return `${prefix}_${index.toString(36)}`;
}

function nextSeed() {
  seed += 13;
  return seed;
}

function esc(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function lines(label) {
  return String(label).split("\n");
}

function textWidth(label, fontSize = 18) {
  return Math.max(...lines(label).map((line) => line.length), 1) * fontSize * 0.55;
}

function textHeight(label, fontSize = 18) {
  return lines(label).length * fontSize * 1.25;
}

function baseElement(type, x, y, width, height, extra = {}) {
  return {
    id: nextId(type),
    type,
    x,
    y,
    width,
    height,
    angle: 0,
    strokeColor: "#1f2937",
    backgroundColor: "transparent",
    fillStyle: "hachure",
    strokeWidth: 2,
    strokeStyle: "solid",
    roughness: 1,
    opacity: 100,
    groupIds: [],
    frameId: null,
    index: `a${index.toString(36)}`,
    roundness: type === "rectangle" ? { type: 3 } : null,
    seed: nextSeed(),
    version: 1,
    versionNonce: nextSeed(),
    isDeleted: false,
    boundElements: null,
    updated: 1,
    link: null,
    locked: false,
    ...extra,
  };
}

function textElement(text, x, y, width, options = {}) {
  const fontSize = options.fontSize ?? 18;
  const height = options.height ?? textHeight(text, fontSize);
  return baseElement("text", x, y, width, height, {
    strokeColor: options.color ?? "#111827",
    backgroundColor: "transparent",
    fillStyle: "solid",
    strokeWidth: 1,
    roughness: 0,
    text,
    fontSize,
    fontFamily: 1,
    textAlign: options.align ?? "center",
    verticalAlign: options.verticalAlign ?? "middle",
    containerId: null,
    originalText: text,
    autoResize: false,
    lineHeight: 1.25,
  });
}

function shapeElement(node) {
  const common = {
    strokeColor: "#1f2937",
    backgroundColor: node.color ?? colors.white,
    fillStyle: "solid",
  };
  if (node.kind === "circle" || node.kind === "db") {
    return baseElement("ellipse", node.x, node.y, node.w, node.h, common);
  }
  if (node.kind === "decision" || node.kind === "diamond") {
    return baseElement("diamond", node.x, node.y, node.w, node.h, common);
  }
  return baseElement("rectangle", node.x, node.y, node.w, node.h, common);
}

function nodeElements(node) {
  const label = textElement(
    node.label,
    node.x + 10,
    node.y + (node.h - textHeight(node.label, node.fontSize ?? 18)) / 2,
    node.w - 20,
    { fontSize: node.fontSize ?? 18 },
  );
  return [shapeElement(node), label];
}

function center(node) {
  return { x: node.x + node.w / 2, y: node.y + node.h / 2 };
}

function boundaryPoint(from, to) {
  const a = center(from);
  const b = center(to);
  const dx = b.x - a.x;
  const dy = b.y - a.y;
  if (Math.abs(dx) * from.h > Math.abs(dy) * from.w) {
    const x = a.x + Math.sign(dx || 1) * from.w / 2;
    const y = a.y + (dy / (Math.abs(dx) || 1)) * from.w / 2;
    return { x, y };
  }
  const y = a.y + Math.sign(dy || 1) * from.h / 2;
  const x = a.x + (dx / (Math.abs(dy) || 1)) * from.h / 2;
  return { x, y };
}

function arrowElement(start, end) {
  return baseElement("arrow", start.x, start.y, end.x - start.x, end.y - start.y, {
    backgroundColor: "transparent",
    fillStyle: "hachure",
    points: [
      [0, 0],
      [end.x - start.x, end.y - start.y],
    ],
    lastCommittedPoint: null,
    startBinding: null,
    endBinding: null,
    startArrowhead: null,
    endArrowhead: "arrow",
  });
}

function svgText(label, x, y, width, options = {}) {
  const fontSize = options.fontSize ?? 18;
  const parts = lines(label);
  const startY = y - ((parts.length - 1) * fontSize * 1.2) / 2;
  return `<text x="${x}" y="${startY}" text-anchor="middle" dominant-baseline="middle" font-family="Virgil, Segoe Print, Comic Sans MS, sans-serif" font-size="${fontSize}" fill="${options.color ?? "#111827"}">${parts
    .map((line, idx) => `<tspan x="${x}" dy="${idx === 0 ? 0 : fontSize * 1.2}">${esc(line)}</tspan>`)
    .join("")}</text>`;
}

function svgNode(node) {
  const fill = node.color ?? colors.white;
  const stroke = "#1f2937";
  const label = svgText(node.label, node.x + node.w / 2, node.y + node.h / 2, node.w, {
    fontSize: node.fontSize ?? 18,
  });
  if (node.kind === "circle") {
    return `<ellipse cx="${node.x + node.w / 2}" cy="${node.y + node.h / 2}" rx="${node.w / 2}" ry="${node.h / 2}" fill="${fill}" stroke="${stroke}" stroke-width="2"/>${label}`;
  }
  if (node.kind === "db") {
    const cx = node.x + node.w / 2;
    return `<path d="M ${node.x} ${node.y + 12} C ${node.x} ${node.y - 4}, ${node.x + node.w} ${node.y - 4}, ${node.x + node.w} ${node.y + 12} L ${node.x + node.w} ${node.y + node.h - 12} C ${node.x + node.w} ${node.y + node.h + 4}, ${node.x} ${node.y + node.h + 4}, ${node.x} ${node.y + node.h - 12} Z M ${node.x} ${node.y + 12} C ${node.x} ${node.y + 28}, ${node.x + node.w} ${node.y + 28}, ${node.x + node.w} ${node.y + 12}" fill="${fill}" stroke="${stroke}" stroke-width="2"/>${label}`;
  }
  if (node.kind === "decision" || node.kind === "diamond") {
    const points = [
      [node.x + node.w / 2, node.y],
      [node.x + node.w, node.y + node.h / 2],
      [node.x + node.w / 2, node.y + node.h],
      [node.x, node.y + node.h / 2],
    ]
      .map((point) => point.join(","))
      .join(" ");
    return `<polygon points="${points}" fill="${fill}" stroke="${stroke}" stroke-width="2"/>${label}`;
  }
  return `<rect x="${node.x}" y="${node.y}" width="${node.w}" height="${node.h}" rx="8" fill="${fill}" stroke="${stroke}" stroke-width="2"/>${label}`;
}

function svgArrow(start, end, label) {
  const midX = (start.x + end.x) / 2;
  const midY = (start.y + end.y) / 2;
  const text = label
    ? `<rect x="${midX - textWidth(label, 13) / 2 - 8}" y="${midY - 15}" width="${textWidth(label, 13) + 16}" height="24" rx="4" fill="#ffffff" opacity="0.92"/>${svgText(label, midX, midY - 1, 240, { fontSize: 13 })}`
    : "";
  return `<line x1="${start.x}" y1="${start.y}" x2="${end.x}" y2="${end.y}" stroke="#1f2937" stroke-width="2" marker-end="url(#arrow)"/>${text}`;
}

function canvas(content, width, height) {
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" role="img">
  <defs>
    <marker id="arrow" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M0,0 L0,6 L9,3 z" fill="#1f2937"/>
    </marker>
  </defs>
  <rect width="100%" height="100%" fill="#fffdf7"/>
  ${content}
</svg>
`;
}

function writeScene(slug, title, elements, width, height, svg) {
  const scene = {
    type: "excalidraw",
    version: 2,
    source: "https://excalidraw.com",
    elements,
    appState: {
      viewBackgroundColor: "#fffdf7",
      currentItemFontFamily: 1,
      exportBackground: true,
      name: title,
    },
    files: {},
  };
  fs.writeFileSync(path.join(__dirname, `${slug}.excalidraw`), `${JSON.stringify(scene, null, 2)}\n`);
  fs.writeFileSync(path.join(__dirname, `${slug}.svg`), svg);
}

function flow(slug, title, width, height, nodes, edges) {
  const byId = new Map(nodes.map((node) => [node.id, node]));
  const elements = [];
  nodes.forEach((node) => elements.push(...nodeElements(node)));
  edges.forEach((edge) => {
    const from = byId.get(edge.from);
    const to = byId.get(edge.to);
    const start = edge.fromPoint ?? boundaryPoint(from, to);
    const end = edge.toPoint ?? boundaryPoint(to, from);
    elements.push(arrowElement(start, end));
    if (edge.label) {
      const midX = (start.x + end.x) / 2;
      const midY = (start.y + end.y) / 2;
      elements.push(textElement(edge.label, midX - 100, midY - 28, 200, { fontSize: 13 }));
    }
  });
  const svg = canvas(
    [
      ...edges.map((edge) => {
        const from = byId.get(edge.from);
        const to = byId.get(edge.to);
        return svgArrow(edge.fromPoint ?? boundaryPoint(from, to), edge.toPoint ?? boundaryPoint(to, from), edge.label);
      }),
      ...nodes.map(svgNode),
    ].join("\n"),
    width,
    height,
  );
  writeScene(slug, title, elements, width, height, svg);
}

function sequence(slug, title, participants, messages) {
  const margin = 70;
  const stepX = 190;
  const top = 45;
  const msgGap = 46;
  const width = margin * 2 + stepX * (participants.length - 1) + 160;
  const height = top + 110 + msgGap * messages.length;
  const xs = new Map(participants.map((p, idx) => [p.id, margin + 80 + idx * stepX]));
  const elements = [];
  let svg = "";
  participants.forEach((p) => {
    const x = xs.get(p.id);
    const node = { id: p.id, label: p.label, x: x - 70, y: top, w: 140, h: 54, color: colors.blue };
    elements.push(...nodeElements(node));
    elements.push(baseElement("line", x, top + 64, 0, height - top - 90, {
      points: [
        [0, 0],
        [0, height - top - 90],
      ],
      strokeStyle: "dashed",
    }));
    svg += `${svgNode(node)}<line x1="${x}" y1="${top + 64}" x2="${x}" y2="${height - 30}" stroke="#6b7280" stroke-width="1.5" stroke-dasharray="6 6"/>\n`;
  });
  messages.forEach((message, idx) => {
    const y = top + 105 + idx * msgGap;
    const fromX = xs.get(message.from);
    const toX = xs.get(message.to);
    const start = { x: fromX, y };
    const end = { x: toX, y };
    elements.push(arrowElement(start, end));
    elements.push(textElement(message.text, Math.min(fromX, toX) + Math.abs(toX - fromX) / 2 - 115, y - 30, 230, { fontSize: 13 }));
    const dash = message.return ? 'stroke-dasharray="6 6"' : "";
    const marker = fromX === toX ? "" : 'marker-end="url(#arrow)"';
    if (fromX === toX) {
      svg += `<path d="M ${fromX} ${y} h 70 v 25 h -70" fill="none" stroke="#1f2937" stroke-width="2" marker-end="url(#arrow)"/>${svgText(message.text, fromX + 80, y + 10, 200, { fontSize: 13 })}\n`;
    } else {
      svg += `<line x1="${fromX}" y1="${y}" x2="${toX}" y2="${y}" stroke="#1f2937" stroke-width="2" ${dash} ${marker}/>${svgText(message.text, (fromX + toX) / 2, y - 16, 240, { fontSize: 13 })}\n`;
    }
  });
  writeScene(slug, title, elements, width, height, canvas(svg, width, height));
}

function stateMachine(slug, title, width, height, states, edges) {
  flow(slug, title, width, height, states, edges);
}

function erDiagram() {
  const tables = [
    {
      id: "products",
      title: "PRODUCTS",
      x: 40,
      y: 60,
      fields: ["uuid id PK", "string name", "string category", "int64 price_in_kurus", "string image_path", "bool available"],
    },
    {
      id: "orders",
      title: "ORDERS",
      x: 410,
      y: 60,
      fields: ["uuid id PK", "string customer_email", "int64 total", "string status", "time created_at"],
    },
    {
      id: "line_items",
      title: "LINE_ITEMS",
      x: 220,
      y: 310,
      fields: ["uuid id PK", "uuid order_id FK", "uuid product_id FK", "int quantity", "int64 price_in_kurus", "string product_name"],
    },
    {
      id: "outbox",
      title: "OUTBOX_EVENTS",
      x: 650,
      y: 280,
      fields: ["uuid id PK", "string event_type", "string aggregate_type", "string aggregate_id", "string routing_key", "text payload", "int attempts", "time published_at"],
    },
  ];
  const width = 940;
  const height = 560;
  const elements = [];
  let svg = "";
  const tableBox = (table) => {
    const h = 46 + table.fields.length * 26;
    const node = { id: table.id, label: table.title, x: table.x, y: table.y, w: 240, h, color: colors.gray, fontSize: 18 };
    elements.push(shapeElement(node));
    elements.push(textElement(table.title, table.x + 12, table.y + 12, 216, { fontSize: 18, height: 24 }));
    svg += `<rect x="${table.x}" y="${table.y}" width="240" height="${h}" rx="8" fill="${colors.gray}" stroke="#1f2937" stroke-width="2"/><line x1="${table.x}" y1="${table.y + 42}" x2="${table.x + 240}" y2="${table.y + 42}" stroke="#1f2937" stroke-width="2"/>${svgText(table.title, table.x + 120, table.y + 23, 220, { fontSize: 18 })}`;
    table.fields.forEach((field, idx) => {
      const fy = table.y + 66 + idx * 26;
      elements.push(textElement(field, table.x + 14, fy - 12, 210, { fontSize: 14, align: "left", height: 20 }));
      svg += `<text x="${table.x + 16}" y="${fy}" font-family="Virgil, Segoe Print, Comic Sans MS, sans-serif" font-size="14" fill="#111827">${esc(field)}</text>`;
    });
    return { ...node, h };
  };
  const boxes = new Map(tables.map((table) => [table.id, tableBox(table)]));
  const relations = [
    ["products", "line_items", "selected by"],
    ["orders", "line_items", "contains"],
    ["orders", "outbox", "emits facts"],
  ];
  relations.forEach(([from, to, label]) => {
    const a = boxes.get(from);
    const b = boxes.get(to);
    const start = boundaryPoint(a, b);
    const end = boundaryPoint(b, a);
    elements.push(arrowElement(start, end));
    elements.push(textElement(label, (start.x + end.x) / 2 - 70, (start.y + end.y) / 2 - 24, 140, { fontSize: 13 }));
    svg += svgArrow(start, end, label);
  });
  writeScene("data-model", "Data Model", elements, width, height, canvas(svg, width, height));
}

function tableDiagram(slug, title, columns, rows, options = {}) {
  const fontSize = options.fontSize ?? 14;
  const headerFontSize = options.headerFontSize ?? 15;
  const cellPaddingX = 12;
  const cellPaddingY = 10;
  const headerHeight = options.headerHeight ?? 42;
  const minRowHeight = options.rowHeight ?? 34;
  const x = options.x ?? 40;
  const y = options.y ?? 40;
  const maxColWidth = options.maxColWidth ?? 280;
  const minColWidth = options.minColWidth ?? 100;
  const rowColors = options.rowColors ?? [colors.white, "#f9fafb"];
  const borderColor = "#1f2937";
  const headerColor = options.headerColor ?? colors.blue;

  const columnWidths = columns.map((column, colIdx) => {
    const values = [column, ...rows.map((row) => row[colIdx] ?? "")];
    const maxWidth = Math.max(...values.map((value) => textWidth(value, fontSize)));
    return Math.max(minColWidth, Math.min(maxColWidth, Math.ceil(maxWidth + cellPaddingX * 2)));
  });

  const rowHeights = rows.map((row) => {
    const tallestCell = Math.max(...row.map((value) => textHeight(value, fontSize)));
    return Math.max(minRowHeight, Math.ceil(tallestCell + cellPaddingY * 2));
  });

  const tableWidth = columnWidths.reduce((sum, width) => sum + width, 0);
  const tableHeight = headerHeight + rowHeights.reduce((sum, height) => sum + height, 0);
  const width = tableWidth + x * 2;
  const height = tableHeight + y * 2;
  const elements = [];
  let svg = "";

  elements.push(baseElement("rectangle", x, y, tableWidth, tableHeight, {
    backgroundColor: colors.white,
    fillStyle: "solid",
  }));
  svg += `<rect x="${x}" y="${y}" width="${tableWidth}" height="${tableHeight}" rx="8" fill="${colors.white}" stroke="${borderColor}" stroke-width="2"/>`;
  svg += `<rect x="${x}" y="${y}" width="${tableWidth}" height="${headerHeight}" rx="8" fill="${headerColor}" stroke="${borderColor}" stroke-width="2"/>`;

  let cursorX = x;
  columns.forEach((column, idx) => {
    const colWidth = columnWidths[idx];
    if (idx > 0) {
      elements.push(baseElement("line", cursorX, y, 0, tableHeight, {
        points: [[0, 0], [0, tableHeight]],
        lastCommittedPoint: null,
      }));
      svg += `<line x1="${cursorX}" y1="${y}" x2="${cursorX}" y2="${y + tableHeight}" stroke="${borderColor}" stroke-width="2"/>`;
    }
    elements.push(textElement(column, cursorX + cellPaddingX, y + 10, colWidth - cellPaddingX * 2, {
      fontSize: headerFontSize,
      align: "left",
      height: headerHeight - 16,
    }));
    svg += `<text x="${cursorX + cellPaddingX}" y="${y + 26}" font-family="Virgil, Segoe Print, Comic Sans MS, sans-serif" font-size="${headerFontSize}" fill="#111827">${esc(column)}</text>`;
    cursorX += colWidth;
  });

  let cursorY = y + headerHeight;
  svg += `<line x1="${x}" y1="${cursorY}" x2="${x + tableWidth}" y2="${cursorY}" stroke="${borderColor}" stroke-width="2"/>`;
  rows.forEach((row, rowIdx) => {
    const rowHeight = rowHeights[rowIdx];
    const fill = rowColors[rowIdx % rowColors.length];
    svg += `<rect x="${x}" y="${cursorY}" width="${tableWidth}" height="${rowHeight}" fill="${fill}" opacity="0.9"/>`;
    let cellX = x;
    row.forEach((value, colIdx) => {
      const colWidth = columnWidths[colIdx];
      elements.push(textElement(String(value), cellX + cellPaddingX, cursorY + cellPaddingY, colWidth - cellPaddingX * 2, {
        fontSize,
        align: "left",
        height: rowHeight - cellPaddingY * 2,
      }));
      const linesForCell = lines(String(value));
      linesForCell.forEach((line, lineIdx) => {
        svg += `<text x="${cellX + cellPaddingX}" y="${cursorY + 22 + lineIdx * (fontSize * 1.2)}" font-family="Virgil, Segoe Print, Comic Sans MS, sans-serif" font-size="${fontSize}" fill="#111827">${esc(line)}</text>`;
      });
      cellX += colWidth;
    });
    cursorY += rowHeight;
    if (rowIdx < rows.length - 1) {
      svg += `<line x1="${x}" y1="${cursorY}" x2="${x + tableWidth}" y2="${cursorY}" stroke="${borderColor}" stroke-width="1.5"/>`;
    }
  });

  writeScene(slug, title, elements, width, height, canvas(svg, width, height));
}

function embed(slug, alt) {
  return `![${alt}](diagrams/${slug}.svg)\n\n[Edit Excalidraw source](diagrams/${slug}.excalidraw)`;
}

function replaceNextMermaid(content, replacement) {
  return content.replace(/```mermaid\n[\s\S]*?\n```/, replacement);
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function replaceTableBetween(content, startMarker, endMarker, replacement) {
  const marker = startMarker;
  const markerIndex = content.indexOf(marker);
  if (markerIndex === -1) {
    return content;
  }

  const searchStart = markerIndex + marker.length;
  const searchEnd = endMarker ? content.indexOf(endMarker, searchStart) : -1;
  const afterMarker = searchEnd === -1 ? content.slice(searchStart) : content.slice(searchStart, searchEnd);
  const tableMatch = afterMarker.match(/\n\n(\|[^\n]*\|\n\|[- :|]+\|\n(?:\|[^\n]*\|\n)+)/);
  if (!tableMatch || tableMatch.index == null) {
    return content;
  }

  const tableStart = searchStart + tableMatch.index + 2;
  const tableEnd = tableStart + tableMatch[1].length;
  return `${content.slice(0, tableStart)}${replacement}\n${content.slice(tableEnd)}`;
}

function updateDocs() {
  const replacements = {
    "architecture.md": (content) => {
      content = content.replace(
        /```mermaid\nC4Context[\s\S]*?\n```\n\nIf a Markdown renderer does not support C4 Mermaid syntax, use this equivalent flowchart:\n\n```mermaid\nflowchart TB[\s\S]*?\n```/,
        embed("system-context", "Coffee Service system context Excalidraw diagram"),
      );
      content = replaceNextMermaid(content, embed("architecture-checkout-sequence", "Checkout sequence Excalidraw diagram"));
      content = replaceNextMermaid(content, embed("architecture-order-state-machine", "Order state machine Excalidraw diagram"));
      content = replaceTableBetween(content, "## Runtime Containers", "## Service Boundaries", embed("architecture-runtime-containers", "Runtime containers table Excalidraw diagram"));
      return content;
    },
    "events.md": (content) => {
      content = replaceNextMermaid(content, embed("event-transport", "Event transport Excalidraw diagram"));
      content = replaceNextMermaid(content, embed("outbox-lifecycle", "Outbox lifecycle Excalidraw diagram"));
      content = replaceTableBetween(content, "## Transport", "## Rules", embed("events-transport-table", "Event transport table Excalidraw diagram"));
      content = replaceTableBetween(content, "## `order.created`", "## `order.status_updated`", embed("events-order-created-fields", "order.created fields table Excalidraw diagram"));
      content = replaceTableBetween(content, "## `order.status_updated`", "## Outbox Lifecycle", embed("events-order-status-updated-fields", "order.status_updated fields table Excalidraw diagram"));
      content = replaceTableBetween(content, "## `password_reset.requested`", "", embed("events-password-reset-fields", "password_reset.requested fields table Excalidraw diagram"));
      return content;
    },
    "api.md": (content) => {
      content = replaceTableBetween(content, "Authenticated application routes use:", "Default local demo accounts:", embed("api-auth-header", "Authentication header table Excalidraw diagram"));
      content = replaceTableBetween(content, "Default local demo accounts:", "## Auth Endpoints", embed("api-demo-accounts", "Demo accounts table Excalidraw diagram"));
      content = replaceTableBetween(content, "## Auth Endpoints", "Login response:", embed("api-auth-endpoints", "Auth endpoints table Excalidraw diagram"));
      content = replaceTableBetween(content, "## Order Endpoints", "## Products", embed("api-order-endpoints", "Order endpoints table Excalidraw diagram"));
      content = replaceNextMermaid(content, embed("api-status-transitions", "Order status transitions Excalidraw diagram"));
      content = replaceTableBetween(content, "Common status codes:", "", embed("api-errors", "Common API errors table Excalidraw diagram"));
      return content;
    },
    "credentials-and-diagrams.md": (content) => {
      content = replaceTableBetween(content, "The frontend signs in through `auth-service` with email/password and then stores a bearer JWT.", "Login request:", embed("credentials-demo-auth", "Demo auth credentials table Excalidraw diagram"));
      content = replaceTableBetween(content, "## Local Infrastructure Credentials", "## Quick Defense Answers", embed("credentials-local-infra", "Local infrastructure credentials table Excalidraw diagram"));
      content = replaceTableBetween(content, "## Quick Defense Answers", "## Database Snapshot", embed("credentials-defense-answers", "Quick defense answers table Excalidraw diagram"));
      [
        ["runtime-system", "Runtime system Excalidraw diagram"],
        ["service-ownership", "Service ownership Excalidraw diagram"],
        ["auth-role-sequence", "Auth and role sequence Excalidraw diagram"],
        ["role-resolution", "Role resolution Excalidraw diagram"],
        ["role-access-matrix", "Role access matrix Excalidraw diagram"],
        ["frontend-workflow", "Frontend workflow Excalidraw diagram"],
        ["checkout-sequence", "Checkout sequence Excalidraw diagram"],
        ["outbox-events", "Outbox and events Excalidraw diagram"],
        ["order-event-routing", "Order event routing Excalidraw diagram"],
        ["notification-flow", "Notification flow Excalidraw diagram"],
        ["order-state-machine", "Order state machine Excalidraw diagram"],
        ["data-model", "Data model Excalidraw diagram"],
        ["future-target-shape", "Future target shape Excalidraw diagram"],
      ].forEach(([slug, alt]) => {
        content = replaceNextMermaid(content, embed(slug, alt));
      });
      return content;
    },
    "runbook.md": (content) => {
      content = replaceTableBetween(content, "Expected endpoints:", "## Smoke Test", embed("runbook-expected-endpoints", "Expected endpoints table Excalidraw diagram"));
      content = replaceTableBetween(content, "Auth service:", "Order service:", embed("runbook-auth-env", "Auth service environment variables table Excalidraw diagram"));
      content = replaceTableBetween(content, "Order service:", "Notification service:", embed("runbook-order-env", "Order service environment variables table Excalidraw diagram"));
      content = replaceTableBetween(content, "Notification service:", "Frontend:", embed("runbook-notification-env", "Notification service environment variables table Excalidraw diagram"));
      content = replaceTableBetween(content, "Frontend:", "## Troubleshooting", embed("runbook-frontend-env", "Frontend environment variables table Excalidraw diagram"));
      content = replaceTableBetween(content, "## Troubleshooting", "## Local Data", embed("runbook-troubleshooting", "Troubleshooting table Excalidraw diagram"));
      return content;
    },
  };
  Object.entries(replacements).forEach(([file, update]) => {
    const fullPath = path.join(docsDir, file);
    fs.writeFileSync(fullPath, update(fs.readFileSync(fullPath, "utf8")));
  });
}

flow(
  "system-context",
  "Coffee Service System Context",
  1100,
  620,
  [
    { id: "customer", label: "Customer", x: 35, y: 135, w: 150, h: 64, color: colors.yellow },
    { id: "staff", label: "Barista / Admin", x: 35, y: 330, w: 150, h: 64, color: colors.yellow },
    { id: "frontend", label: "React frontend\nVite + Nginx", x: 255, y: 225, w: 180, h: 84, color: colors.blue },
    { id: "auth", label: "Auth service\nemail/password + JWT", x: 520, y: 70, w: 200, h: 84, color: colors.purple },
    { id: "order", label: "Order service\nproducts + orders", x: 520, y: 250, w: 200, h: 84, color: colors.green },
    { id: "db", label: "PostgreSQL\nshared runtime DB", x: 820, y: 150, w: 190, h: 88, kind: "db", color: colors.gray },
    { id: "rabbit", label: "RabbitMQ\ncoffee.auth + coffee.orders", x: 800, y: 330, w: 220, h: 92, kind: "db", color: colors.orange, fontSize: 16 },
    { id: "notify", label: "Notification\nservice", x: 520, y: 465, w: 190, h: 82, color: colors.green },
    { id: "mailhog", label: "MailHog\nSMTP capture", x: 820, y: 470, w: 190, h: 82, color: colors.red },
  ],
  [
    { from: "customer", to: "frontend", label: "uses" },
    { from: "staff", to: "frontend", label: "uses" },
    { from: "frontend", to: "auth", label: "POST /auth/login" },
    { from: "frontend", to: "order", label: "HTTP JSON + Bearer JWT" },
    { from: "auth", to: "db", label: "users + outbox" },
    { from: "order", to: "db", label: "products + orders + outbox" },
    { from: "auth", to: "rabbit", label: "coffee.auth" },
    { from: "order", to: "rabbit", label: "coffee.orders" },
    { from: "rabbit", to: "notify", label: "deliver facts" },
    { from: "notify", to: "mailhog", label: "SMTP" },
  ],
);

sequence("architecture-checkout-sequence", "Checkout Sequence", [
  { id: "ui", label: "React\nfrontend" },
  { id: "api", label: "Order\nservice" },
  { id: "db", label: "PostgreSQL" },
  { id: "outbox", label: "Outbox\ndispatcher" },
  { id: "mq", label: "RabbitMQ" },
  { id: "notify", label: "Notification\nservice" },
  { id: "smtp", label: "MailHog /\nSMTP" },
], [
  { from: "ui", to: "api", text: "POST /orders" },
  { from: "api", to: "db", text: "Lookup products" },
  { from: "api", to: "db", text: "Create order + outbox" },
  { from: "api", to: "ui", text: "201 OrderResponse", return: true },
  { from: "outbox", to: "db", text: "Read unpublished" },
  { from: "outbox", to: "mq", text: "Publish order.created" },
  { from: "outbox", to: "db", text: "Mark published" },
  { from: "mq", to: "notify", text: "Deliver event" },
  { from: "notify", to: "smtp", text: "Send receipt email" },
  { from: "notify", to: "mq", text: "Ack", return: true },
]);

const orderStates = [
  { id: "start", label: "start", x: 50, y: 145, w: 70, h: 70, kind: "circle", color: colors.gray },
  { id: "preparing", label: "preparing", x: 200, y: 140, w: 150, h: 80, color: colors.yellow },
  { id: "ready", label: "ready", x: 450, y: 140, w: 130, h: 80, color: colors.green },
  { id: "completed", label: "completed", x: 700, y: 80, w: 150, h: 80, color: colors.blue },
  { id: "cancelled", label: "cancelled", x: 700, y: 240, w: 150, h: 80, color: colors.red },
];
const orderEdges = [
  { from: "start", to: "preparing" },
  { from: "preparing", to: "ready", label: "/ready" },
  { from: "ready", to: "completed", label: "/complete" },
  { from: "preparing", to: "cancelled", label: "/cancel" },
  { from: "ready", to: "cancelled", label: "/cancel" },
];
stateMachine("architecture-order-state-machine", "Order State Machine", 900, 390, orderStates, orderEdges);

flow(
  "event-transport",
  "Event Transport",
  980,
  230,
  [
    { id: "db", label: "order-service DB", x: 35, y: 70, w: 150, h: 70, kind: "db", color: colors.gray },
    { id: "outbox", label: "outbox_events", x: 230, y: 70, w: 150, h: 70, color: colors.yellow },
    { id: "dispatch", label: "Outbox\ndispatcher", x: 425, y: 65, w: 150, h: 80, color: colors.green },
    { id: "exchange", label: "RabbitMQ topic\ncoffee.orders", x: 620, y: 60, w: 170, h: 90, kind: "db", color: colors.orange },
    { id: "queue", label: "notification-service.orders", x: 825, y: 70, w: 130, h: 70, color: colors.blue, fontSize: 14 },
  ],
  [
    { from: "db", to: "outbox" },
    { from: "outbox", to: "dispatch" },
    { from: "dispatch", to: "exchange" },
    { from: "exchange", to: "queue" },
  ],
);

sequence("outbox-lifecycle", "Outbox Lifecycle", [
  { id: "service", label: "Order\nservice" },
  { id: "db", label: "PostgreSQL" },
  { id: "worker", label: "Outbox\ndispatcher" },
  { id: "mq", label: "RabbitMQ" },
], [
  { from: "service", to: "db", text: "Begin transaction" },
  { from: "service", to: "db", text: "Write order/status" },
  { from: "service", to: "db", text: "Insert outbox event" },
  { from: "service", to: "db", text: "Commit" },
  { from: "worker", to: "db", text: "Load unpublished rows" },
  { from: "worker", to: "mq", text: "Publish event" },
  { from: "mq", to: "worker", text: "Publish accepted", return: true },
  { from: "worker", to: "db", text: "Mark published_at" },
]);

stateMachine("api-status-transitions", "API Status Transitions", 900, 390, orderStates, orderEdges);

flow(
  "runtime-system",
  "Runtime System",
  1120,
  620,
  [
    { id: "browser", label: "Browser", x: 35, y: 255, w: 130, h: 70, color: colors.yellow },
    { id: "frontend", label: "Frontend\nNginx :80", x: 225, y: 245, w: 160, h: 90, color: colors.blue },
    { id: "auth", label: "Auth service\nGo/Gin :8081", x: 475, y: 80, w: 185, h: 90, color: colors.purple },
    { id: "order", label: "Order service\nGo/Gin :8080", x: 475, y: 245, w: 185, h: 90, color: colors.green },
    { id: "db", label: "PostgreSQL\n:5432", x: 780, y: 155, w: 165, h: 78, kind: "db", color: colors.gray },
    { id: "mq", label: "RabbitMQ\n:5672 / :15672", x: 780, y: 340, w: 180, h: 92, kind: "db", color: colors.orange },
    { id: "notify", label: "Notification\nservice", x: 475, y: 470, w: 185, h: 80, color: colors.green },
    { id: "mailhog", label: "MailHog\nSMTP :1025\nUI :8025", x: 780, y: 475, w: 180, h: 92, color: colors.red },
  ],
  [
    { from: "browser", to: "frontend", label: "loads app" },
    { from: "frontend", to: "auth", label: "POST /auth/login" },
    { from: "frontend", to: "order", label: "Bearer JWT API calls" },
    { from: "auth", to: "db", label: "users + outbox" },
    { from: "order", to: "db", label: "products + orders + outbox" },
    { from: "auth", to: "mq", label: "coffee.auth" },
    { from: "order", to: "mq", label: "coffee.orders" },
    { from: "mq", to: "notify", label: "deliver facts" },
    { from: "notify", to: "mailhog", label: "SMTP" },
  ],
);

flow(
  "service-ownership",
  "Service Ownership",
  1020,
  520,
  [
    { id: "ui", label: "Frontend\nlogin, menu, cart,\norders, queue", x: 40, y: 210, w: 190, h: 100, color: colors.blue },
    { id: "auth", label: "Auth service\nusers + JWT issue", x: 320, y: 60, w: 190, h: 84, color: colors.purple },
    { id: "products", label: "Order service\nproducts", x: 320, y: 190, w: 190, h: 84, color: colors.green },
    { id: "orders", label: "Order service\norders + workflow", x: 320, y: 320, w: 190, h: 84, color: colors.green },
    { id: "outbox", label: "Transactional\noutboxes", x: 610, y: 190, w: 170, h: 84, color: colors.yellow },
    { id: "consumer", label: "Notification service\nRabbitMQ consumer", x: 610, y: 330, w: 190, h: 84, color: colors.orange },
    { id: "email", label: "Email formatting\nSMTP / log sender", x: 830, y: 255, w: 160, h: 84, color: colors.red, fontSize: 16 },
  ],
  [
    { from: "ui", to: "auth" },
    { from: "ui", to: "products" },
    { from: "ui", to: "orders" },
    { from: "auth", to: "outbox" },
    { from: "orders", to: "outbox" },
    { from: "outbox", to: "consumer" },
    { from: "consumer", to: "email" },
  ],
);

sequence("auth-role-sequence", "Auth And Role Flow", [
  { id: "user", label: "Browser\nuser" },
  { id: "ui", label: "React\nfrontend" },
  { id: "auth", label: "Auth\nservice" },
  { id: "api", label: "Order\nservice" },
], [
  { from: "user", to: "ui", text: "Choose demo account" },
  { from: "ui", to: "auth", text: "POST /auth/login with JSON email/password" },
  { from: "auth", to: "ui", text: "Return Bearer JWT", return: true },
  { from: "ui", to: "api", text: "HTTP request with JWT" },
  { from: "api", to: "api", text: "Verify token + check role" },
  { from: "api", to: "ui", text: "JSON response", return: true },
]);

flow(
  "role-resolution",
  "Role Resolution",
  920,
  570,
  [
    { id: "req", label: "HTTP request", x: 40, y: 235, w: 140, h: 70, color: colors.blue },
    { id: "session", label: "Bearer JWT\npresent?", x: 260, y: 215, w: 130, h: 110, kind: "decision", color: colors.yellow },
    { id: "guest", label: "401\nrequired", x: 500, y: 340, w: 140, h: 70, color: colors.gray },
    { id: "email", label: "Verify token\nand claims", x: 500, y: 170, w: 140, h: 80, color: colors.green },
    { id: "adminq", label: "Role is\nadmin?", x: 700, y: 145, w: 120, h: 100, kind: "decision", color: colors.yellow },
    { id: "admin", label: "Role: admin", x: 740, y: 25, w: 140, h: 70, color: colors.purple },
    { id: "staffq", label: "Role is\nbarista?", x: 700, y: 285, w: 120, h: 100, kind: "decision", color: colors.yellow },
    { id: "staff", label: "Role: barista", x: 740, y: 430, w: 140, h: 70, color: colors.orange },
    { id: "user", label: "Role: user", x: 520, y: 475, w: 140, h: 70, color: colors.green },
    { id: "guard", label: "Handler\nrole guard", x: 305, y: 455, w: 150, h: 80, color: colors.red },
  ],
  [
    { from: "req", to: "session" },
    { from: "session", to: "guest", label: "missing" },
    { from: "session", to: "email", label: "present" },
    { from: "email", to: "adminq" },
    { from: "adminq", to: "admin", label: "Yes" },
    { from: "adminq", to: "staffq", label: "No" },
    { from: "staffq", to: "staff", label: "Yes" },
    { from: "staffq", to: "user", label: "No" },
    { from: "guest", to: "guard" },
    { from: "admin", to: "guard" },
    { from: "staff", to: "guard" },
    { from: "user", to: "guard" },
  ],
);

flow(
  "role-access-matrix",
  "Role Access Matrix",
  980,
  540,
  [
    { id: "user", label: "user", x: 60, y: 70, w: 120, h: 60, color: colors.green },
    { id: "barista", label: "barista", x: 60, y: 235, w: 120, h: 60, color: colors.orange },
    { id: "admin", label: "admin", x: 60, y: 400, w: 120, h: 60, color: colors.purple },
    { id: "browse", label: "GET\n/products", x: 310, y: 45, w: 150, h: 70, color: colors.blue },
    { id: "create", label: "POST\n/orders", x: 520, y: 45, w: 150, h: 70, color: colors.blue },
    { id: "mine", label: "GET\n/orders/mine", x: 730, y: 45, w: 160, h: 70, color: colors.blue },
    { id: "queue", label: "GET\n/staff/orders", x: 310, y: 355, w: 150, h: 70, color: colors.yellow },
    { id: "status", label: "POST staff\nstatus actions", x: 520, y: 355, w: 170, h: 70, color: colors.yellow },
  ],
  [
    { from: "user", to: "browse" },
    { from: "user", to: "create" },
    { from: "user", to: "mine" },
    { from: "barista", to: "browse" },
    { from: "barista", to: "queue" },
    { from: "barista", to: "status" },
    { from: "admin", to: "browse" },
    { from: "admin", to: "create" },
    { from: "admin", to: "mine" },
    { from: "admin", to: "queue" },
    { from: "admin", to: "status" },
  ],
);

flow(
  "frontend-workflow",
  "Frontend Workflow",
  980,
  560,
  [
    { id: "start", label: "Open\nfrontend", x: 40, y: 240, w: 130, h: 70, color: colors.yellow },
    { id: "auth", label: "Login panel\n/demo creds", x: 230, y: 240, w: 150, h: 70, color: colors.purple },
    { id: "me", label: "POST /auth/login\nstore JWT", x: 430, y: 240, w: 130, h: 70, color: colors.purple },
    { id: "load", label: "Load\n/products", x: 620, y: 85, w: 130, h: 70, color: colors.blue },
    { id: "menu", label: "Menu\nview", x: 815, y: 85, w: 130, h: 70, color: colors.green },
    { id: "cart", label: "Add products\nto cart", x: 815, y: 200, w: 150, h: 70, color: colors.green },
    { id: "checkout", label: "Submit\n/orders", x: 815, y: 315, w: 130, h: 70, color: colors.orange },
    { id: "orders", label: "Orders\nview", x: 815, y: 430, w: 130, h: 70, color: colors.blue },
    { id: "mine", label: "Load\n/orders/mine", x: 620, y: 430, w: 150, h: 70, color: colors.blue },
    { id: "role", label: "Role?", x: 620, y: 420, w: 120, h: 90, kind: "decision", color: colors.yellow },
    { id: "staff", label: "Barista\nqueue", x: 430, y: 430, w: 130, h: 70, color: colors.orange },
  ],
  [
    { from: "start", to: "auth" },
    { from: "auth", to: "me" },
    { from: "me", to: "load" },
    { from: "load", to: "menu" },
    { from: "menu", to: "cart" },
    { from: "cart", to: "checkout" },
    { from: "checkout", to: "orders" },
    { from: "orders", to: "mine" },
    { from: "me", to: "role" },
    { from: "role", to: "menu", label: "user/admin" },
    { from: "role", to: "staff", label: "barista/admin" },
    { from: "staff", to: "orders", label: "status action" },
  ],
);

sequence("checkout-sequence", "Checkout Sequence", [
  { id: "ui", label: "React\nfrontend" },
  { id: "handler", label: "Order\nhandler" },
  { id: "service", label: "Order\nservice" },
  { id: "db", label: "PostgreSQL" },
  { id: "outbox", label: "Outbox\nrow" },
], [
  { from: "ui", to: "handler", text: "POST /orders" },
  { from: "handler", to: "handler", text: "Use JWT email / request email" },
  { from: "handler", to: "service", text: "CreateOrder(input)" },
  { from: "service", to: "db", text: "Lookup products" },
  { from: "service", to: "service", text: "Calculate trusted total" },
  { from: "service", to: "db", text: "Insert order + items" },
  { from: "service", to: "outbox", text: "Insert order.created" },
  { from: "db", to: "service", text: "Commit transaction", return: true },
  { from: "service", to: "handler", text: "Order model", return: true },
  { from: "handler", to: "ui", text: "201 OrderResponse", return: true },
]);

sequence("outbox-events", "Outbox And Events", [
  { id: "order", label: "Order\ntransaction" },
  { id: "db", label: "outbox_events" },
  { id: "dispatcher", label: "Outbox\ndispatcher" },
  { id: "mq", label: "coffee.orders" },
  { id: "consumer", label: "Notification\nservice" },
], [
  { from: "order", to: "db", text: "Write change + event atomically" },
  { from: "dispatcher", to: "db", text: "Poll unpublished events" },
  { from: "dispatcher", to: "mq", text: "Publish routing key" },
  { from: "dispatcher", to: "db", text: "Set published_at / error" },
  { from: "mq", to: "consumer", text: "Deliver event" },
  { from: "consumer", to: "mq", text: "Ack after handling", return: true },
]);

flow(
  "order-event-routing",
  "Order Event Routing",
  850,
  280,
  [
    { id: "outbox", label: "Order service\noutbox", x: 40, y: 100, w: 160, h: 80, color: colors.yellow },
    { id: "exchange", label: "coffee.orders\ntopic exchange", x: 285, y: 95, w: 170, h: 90, kind: "db", color: colors.orange },
    { id: "queue", label: "notification-service.orders", x: 550, y: 105, w: 170, h: 70, color: colors.blue, fontSize: 15 },
    { id: "consumer", label: "Notification\nconsumer", x: 550, y: 200, w: 170, h: 70, color: colors.green },
    { id: "mail", label: "Receipt /\nstatus email", x: 315, y: 200, w: 160, h: 70, color: colors.red },
  ],
  [
    { from: "outbox", to: "exchange" },
    { from: "exchange", to: "queue", label: "order.created" },
    { from: "exchange", to: "queue", label: "order.status_updated", fromPoint: { x: 455, y: 160 }, toPoint: { x: 550, y: 160 } },
    { from: "queue", to: "consumer" },
    { from: "consumer", to: "mail" },
  ],
);

flow(
  "notification-flow",
  "Notification Flow",
  980,
  620,
  [
    { id: "event", label: "Order event", x: 40, y: 260, w: 140, h: 70, color: colors.blue },
    { id: "known", label: "Known\nevent type?", x: 250, y: 235, w: 130, h: 110, kind: "decision", color: colors.yellow },
    { id: "ackUnknown", label: "Ack and\nignore", x: 470, y: 105, w: 140, h: 70, color: colors.gray },
    { id: "dup", label: "Event already\nseen?", x: 470, y: 235, w: 140, h: 110, kind: "decision", color: colors.yellow },
    { id: "ackDup", label: "Ack\nduplicate", x: 690, y: 105, w: 140, h: 70, color: colors.gray },
    { id: "format", label: "Format\ncustomer email", x: 690, y: 255, w: 150, h: 80, color: colors.green },
    { id: "send", label: "Send via SMTP\nor log sender", x: 690, y: 400, w: 160, h: 80, color: colors.orange },
    { id: "success", label: "Send\nsucceeded?", x: 470, y: 385, w: 140, h: 110, kind: "decision", color: colors.yellow },
    { id: "remember", label: "Remember\nevent id", x: 250, y: 395, w: 140, h: 70, color: colors.green },
    { id: "ack", label: "Ack\nmessage", x: 40, y: 395, w: 140, h: 70, color: colors.blue },
    { id: "nack", label: "Nack /\nretry path", x: 470, y: 535, w: 140, h: 70, color: colors.red },
  ],
  [
    { from: "event", to: "known" },
    { from: "known", to: "ackUnknown", label: "No" },
    { from: "known", to: "dup", label: "Yes" },
    { from: "dup", to: "ackDup", label: "Yes" },
    { from: "dup", to: "format", label: "No" },
    { from: "format", to: "send" },
    { from: "send", to: "success" },
    { from: "success", to: "remember", label: "Yes" },
    { from: "remember", to: "ack" },
    { from: "success", to: "nack", label: "No" },
  ],
);

stateMachine("order-state-machine", "Order State Machine", 900, 390, orderStates, orderEdges);
erDiagram();
tableDiagram(
  "architecture-runtime-containers",
  "Runtime Containers",
  ["Container", "Purpose"],
  [
    ["frontend", "Serves the Vite-built React console through Nginx."],
    ["auth-service", "Owns users, password hashes, JWT issuance, role identity, and auth outbox events."],
    ["order-service", "Owns products, orders, checkout, status workflow, and order outbox events."],
    ["notification-service", "Consumes order/auth facts and sends emails. It does not write to another service database."],
    ["postgres", "Shared PostgreSQL instance with service-owned tables."],
    ["rabbitmq", "Topic exchange transport for service facts."],
    ["mailhog", "Local SMTP sink and email inspection UI."],
  ],
);
tableDiagram(
  "api-auth-header",
  "Authentication Header",
  ["Header", "Values", "Purpose"],
  [["Authorization", "Bearer <jwt>", "Carries the signed session token issued by auth-service."]],
  { maxColWidth: 300 },
);
tableDiagram(
  "api-demo-accounts",
  "Demo Accounts",
  ["Email", "Password", "Role"],
  [
    ["customer@example.com", "customer123", "user"],
    ["barista@coffee.local", "barista123", "barista"],
    ["admin@coffee.local", "admin123", "admin"],
  ],
);
tableDiagram(
  "api-auth-endpoints",
  "Auth Endpoints",
  ["Method", "Path", "Role", "Description"],
  [
    ["GET", "/ping", "public", "Auth health check."],
    ["POST", "/auth/login", "public", "Exchanges email/password for a JWT."],
    ["GET", "/auth/me", "authenticated", "Returns the current token subject, email, and role."],
    ["POST", "/auth/password-reset-requests", "public", "Enqueues password_reset.requested if the email exists. Always returns 202."],
  ],
  { maxColWidth: 340 },
);
tableDiagram(
  "api-order-endpoints",
  "Order Endpoints",
  ["Method", "Path", "Role", "Description"],
  [
    ["GET", "/ping", "public", "Order health check."],
    ["GET", "/products", "user, barista, admin", "Lists all menu products."],
    ["POST", "/orders", "user, admin", "Creates an order and enqueues order.created."],
    ["GET", "/orders/mine", "user, admin", "Lists orders for the JWT email. Query email is only a local fallback."],
    ["GET", "/staff/orders", "barista, admin", "Lists all orders for the staff queue."],
    ["POST", "/staff/orders/:id/ready", "barista, admin", "Moves preparing to ready and enqueues order.status_updated."],
    ["POST", "/staff/orders/:id/complete", "barista, admin", "Moves ready to completed and enqueues order.status_updated."],
    ["POST", "/staff/orders/:id/cancel", "barista, admin", "Cancels preparing or ready and enqueues order.status_updated."],
  ],
  { maxColWidth: 330 },
);
tableDiagram(
  "api-errors",
  "Common API Errors",
  ["Status", "Meaning"],
  [
    ["400", "Invalid request, validation error, missing customer email, invalid transition, or unknown product."],
    ["401", "Missing or invalid email-password login result or missing or invalid JWT."],
    ["403", "Authenticated role is not allowed."],
    ["404", "Resource not found."],
    ["500", "Unexpected service or database failure."],
  ],
);
tableDiagram(
  "events-transport-table",
  "Event Transport",
  ["Exchange", "Routing keys", "Publisher", "Consumer"],
  [
    ["coffee.orders", "order.created\norder.status_updated", "order-service", "notification-service"],
    ["coffee.auth", "password_reset.requested", "auth-service", "notification-service"],
  ],
  { maxColWidth: 250 },
);
tableDiagram(
  "events-order-created-fields",
  "order.created Fields",
  ["Field", "Type", "Notes"],
  [
    ["event_id", "string UUID", "Idempotency key."],
    ["order_id", "string UUID", "Aggregate identifier."],
    ["customer_email", "string", "Receipt destination."],
    ["status", "string", "Initial order status."],
    ["items", "array", "Product snapshot at checkout time."],
    ["total", "integer", "Total in kurus."],
    ["occurred_at", "timestamp", "UTC event time."],
  ],
);
tableDiagram(
  "events-order-status-updated-fields",
  "order.status_updated Fields",
  ["Field", "Type", "Notes"],
  [
    ["event_id", "string UUID", "Idempotency key."],
    ["order_id", "string UUID", "Aggregate identifier."],
    ["customer_email", "string", "Notification destination."],
    ["previous_status", "string", "Status before transition."],
    ["status", "string", "New status."],
    ["occurred_at", "timestamp", "UTC event time."],
  ],
);
tableDiagram(
  "events-password-reset-fields",
  "password_reset.requested Fields",
  ["Field", "Type", "Notes"],
  [
    ["event_id", "string UUID", "Idempotency key."],
    ["user_id", "string UUID", "Aggregate identifier."],
    ["email", "string", "Notification destination."],
    ["role", "string", "Current role snapshot."],
    ["requested_at", "timestamp", "UTC request time."],
  ],
);
tableDiagram(
  "credentials-demo-auth",
  "Demo Auth",
  ["Role", "Email", "Password"],
  [
    ["User", "customer@example.com", "customer123"],
    ["Barista", "barista@coffee.local", "barista123"],
    ["Admin", "admin@coffee.local", "admin123"],
  ],
);
tableDiagram(
  "credentials-local-infra",
  "Local Infrastructure Credentials",
  ["Component", "URL or DSN", "Username", "Password", "Notes"],
  [
    ["Frontend", "http://localhost", "None", "None", "React console served by Nginx through Traefik."],
    ["Traefik gateway", "http://localhost", "None", "None", "Primary entrypoint for frontend and APIs."],
    ["Auth API", "http://localhost/auth", "Demo credentials above", "Demo credentials above", "Login and token-backed identity."],
    ["Order API", "http://localhost/api", "Bearer JWT", "Bearer JWT", "Product and order API."],
    ["PostgreSQL", "postgres://postgres:postgres@localhost:5432/coffee", "postgres", "postgres", "Shared instance; auth and order own separate tables."],
    ["RabbitMQ AMQP", "amqp://guest:guest@localhost:5672/", "guest", "guest", "Used by both outbox dispatchers and notification-service."],
    ["RabbitMQ UI", "http://localhost:15672", "guest", "guest", "Management UI."],
    ["MailHog UI", "http://localhost:8025", "None", "None", "Local email inspection UI."],
    ["MailHog SMTP", "localhost:1025", "None", "None", "Notification-service sends here locally."],
  ],
  { maxColWidth: 320 },
);
tableDiagram(
  "credentials-defense-answers",
  "Quick Defense Answers",
  ["Question", "Answer"],
  [
    ["How many services?", "Three application services: auth-service, order-service, notification-service."],
    ["How many APIs?", "Two HTTP APIs: auth-service and order-service. Notification-service is event-only."],
    ["Why split auth?", "Email/password auth now has its own user table, JWT boundary, and auth events without mixing that logic into order ownership."],
    ["Which roles exist?", "user, barista, admin."],
    ["Why RabbitMQ?", "It decouples order and auth side effects from email delivery and lets notification handling retry independently."],
  ],
);
tableDiagram(
  "runbook-expected-endpoints",
  "Expected Endpoints",
  ["Component", "URL"],
  [
    ["Frontend", "http://localhost"],
    ["Gateway auth route", "http://localhost/auth"],
    ["Gateway order route", "http://localhost/api"],
    ["Auth health", "http://localhost:8081/ping"],
    ["Order health", "http://localhost:8080/ping"],
    ["RabbitMQ management", "http://localhost:15672"],
    ["MailHog", "http://localhost:8025"],
  ],
);
tableDiagram(
  "runbook-auth-env",
  "Auth Service Environment Variables",
  ["Variable", "Default", "Purpose"],
  [
    ["PORT", "8081", "HTTP server port."],
    ["AUTH_DB_URL", "local Postgres DSN", "Auth database connection."],
    ["RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/", "RabbitMQ connection for auth outbox dispatch."],
    ["API_DOMAIN", "http://localhost", "Order API origin allowed by CORS."],
    ["AUTH_API_DOMAIN", "http://localhost", "Auth API origin allowed by CORS."],
    ["WEBSITE_DOMAIN", "http://localhost", "Frontend origin allowed by CORS."],
    ["JWT_SECRET", "coffee-service-local-jwt-secret", "HMAC secret for signed bearer tokens."],
    ["JWT_ISSUER", "coffee-service", "JWT issuer claim value."],
    ["JWT_TTL_MINUTES", "480", "Token lifetime in minutes."],
    ["AUTH_DEMO_USERS", "built-in user/barista/admin accounts", "Comma-separated email:password:role entries."],
  ],
  { maxColWidth: 320 },
);
tableDiagram(
  "runbook-order-env",
  "Order Service Environment Variables",
  ["Variable", "Default", "Purpose"],
  [
    ["PORT", "8080", "HTTP server port."],
    ["DB_URL", "local Postgres DSN", "Runtime database connection."],
    ["RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/", "RabbitMQ connection."],
    ["API_DOMAIN", "http://localhost", "API origin allowed by CORS."],
    ["AUTH_API_DOMAIN", "http://localhost", "Auth API origin allowed by CORS."],
    ["WEBSITE_DOMAIN", "http://localhost", "Frontend origin allowed by CORS."],
    ["JWT_SECRET", "coffee-service-local-jwt-secret", "HMAC secret for signed bearer tokens."],
    ["JWT_ISSUER", "coffee-service", "JWT issuer claim value."],
    ["JWT_TTL_MINUTES", "480", "Token lifetime in minutes."],
  ],
  { maxColWidth: 320 },
);
tableDiagram(
  "runbook-notification-env",
  "Notification Service Environment Variables",
  ["Variable", "Default", "Purpose"],
  [
    ["RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/", "RabbitMQ connection."],
    ["SMTP_HOST", "mailhog", "SMTP server host. Empty means log-only sender."],
    ["SMTP_PORT", "1025", "SMTP port."],
    ["SMTP_USERNAME", "empty", "Optional SMTP username."],
    ["SMTP_PASSWORD", "empty", "Optional SMTP password."],
    ["SMTP_FROM", "Coffee Service <orders@coffee.local>", "Sender address."],
    ["NOTIFICATION_FALLBACK_EMAIL", "dev@coffee.local", "Fallback when an event has no customer email."],
  ],
  { maxColWidth: 320 },
);
tableDiagram(
  "runbook-frontend-env",
  "Frontend Environment Variables",
  ["Variable", "Default", "Purpose"],
  [
    ["VITE_API_URL", "http://localhost/api", "Order API base URL used by the browser."],
    ["VITE_AUTH_API_URL", "http://localhost", "Auth API base URL used by the browser."],
  ],
);
tableDiagram(
  "runbook-troubleshooting",
  "Troubleshooting",
  ["Symptom", "Check"],
  [
    ["Staff queue returns 403", "Log in with the barista or admin demo account, then retry the queue request with the bearer token."],
    ["Products return 401", "Log in again so the frontend stores a fresh JWT, or send Authorization: Bearer <token>."],
    ["Login returns 401", "Confirm you are calling http://localhost/auth/login with JSON email and password."],
    ["Orders are created but no email appears", "Check RabbitMQ health, notification-service logs, and MailHog at http://localhost:8025."],
    ["Product list is empty", "Check order-service startup logs for migration or seed errors."],
    ["make check fails at Docker Compose config", "Run docker compose version and confirm Docker is installed."],
  ],
  { maxColWidth: 360 },
);

flow(
  "future-target-shape",
  "Future Target Shape",
  980,
  390,
  [
    { id: "browser", label: "Browser", x: 40, y: 150, w: 130, h: 70, color: colors.yellow },
    { id: "gateway", label: "Future Go\ngateway", x: 245, y: 145, w: 150, h: 80, color: colors.blue },
    { id: "auth", label: "Authentik\nOIDC/JWKS", x: 245, y: 35, w: 150, h: 75, color: colors.purple },
    { id: "order", label: "Order\nservice", x: 485, y: 95, w: 150, h: 75, color: colors.green },
    { id: "product", label: "Future product\nservice", x: 485, y: 235, w: 160, h: 75, color: colors.green },
    { id: "orderdb", label: "Order DB", x: 735, y: 95, w: 130, h: 70, kind: "db", color: colors.gray },
    { id: "productdb", label: "Product DB", x: 735, y: 235, w: 130, h: 70, kind: "db", color: colors.gray },
    { id: "mq", label: "RabbitMQ", x: 735, y: 15, w: 130, h: 70, kind: "db", color: colors.orange },
    { id: "notify", label: "Notification\nservice", x: 805, y: 315, w: 150, h: 65, color: colors.red },
  ],
  [
    { from: "browser", to: "gateway" },
    { from: "gateway", to: "auth", label: "validate JWT" },
    { from: "gateway", to: "order", label: "identity headers" },
    { from: "gateway", to: "product" },
    { from: "order", to: "orderdb" },
    { from: "product", to: "productdb" },
    { from: "order", to: "mq" },
    { from: "mq", to: "notify" },
  ],
);

updateDocs();
