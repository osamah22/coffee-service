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

function embed(slug, alt) {
  return `![${alt}](diagrams/${slug}.svg)\n\n[Edit Excalidraw source](diagrams/${slug}.excalidraw)`;
}

function replaceNextMermaid(content, replacement) {
  return content.replace(/```mermaid\n[\s\S]*?\n```/, replacement);
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
      return content;
    },
    "events.md": (content) => {
      content = replaceNextMermaid(content, embed("event-transport", "Event transport Excalidraw diagram"));
      content = replaceNextMermaid(content, embed("outbox-lifecycle", "Outbox lifecycle Excalidraw diagram"));
      return content;
    },
    "api.md": (content) => replaceNextMermaid(content, embed("api-status-transitions", "Order status transitions Excalidraw diagram")),
    "credentials-and-diagrams.md": (content) => {
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
  };
  Object.entries(replacements).forEach(([file, update]) => {
    const fullPath = path.join(docsDir, file);
    fs.writeFileSync(fullPath, update(fs.readFileSync(fullPath, "utf8")));
  });
}

flow(
  "system-context",
  "Coffee Service System Context",
  980,
  560,
  [
    { id: "customer", label: "Customer", x: 40, y: 120, w: 140, h: 64, color: colors.yellow },
    { id: "staff", label: "Staff / Admin", x: 40, y: 300, w: 140, h: 64, color: colors.yellow },
    { id: "frontend", label: "React frontend\nVite + Nginx", x: 260, y: 200, w: 170, h: 80, color: colors.blue },
    { id: "order", label: "Order service\nGo / Gin / GORM", x: 520, y: 190, w: 180, h: 90, color: colors.green },
    { id: "auth", label: "Basic login\nJWT claims", x: 510, y: 40, w: 170, h: 70, color: colors.purple },
    { id: "db", label: "PostgreSQL\nruntime DB", x: 760, y: 60, w: 160, h: 78, kind: "db", color: colors.gray },
    { id: "rabbit", label: "RabbitMQ\ncoffee.orders", x: 760, y: 230, w: 160, h: 78, kind: "db", color: colors.orange },
    { id: "notify", label: "Notification\nservice", x: 520, y: 390, w: 180, h: 80, color: colors.green },
    { id: "mailhog", label: "MailHog\nSMTP capture", x: 760, y: 390, w: 160, h: 78, color: colors.red },
  ],
  [
    { from: "customer", to: "frontend", label: "uses" },
    { from: "staff", to: "frontend", label: "uses" },
    { from: "frontend", to: "order", label: "HTTP JSON + Bearer JWT" },
    { from: "order", to: "auth", label: "role claims" },
    { from: "order", to: "db", label: "GORM" },
    { from: "order", to: "rabbit", label: "publish events" },
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
  980,
  560,
  [
    { id: "browser", label: "Browser", x: 35, y: 235, w: 130, h: 70, color: colors.yellow },
    { id: "frontend", label: "React frontend\nNginx :80", x: 245, y: 225, w: 170, h: 90, color: colors.blue },
    { id: "api", label: "Order service\nGo/Gin :8080", x: 500, y: 225, w: 180, h: 90, color: colors.green },
    { id: "st", label: "Basic login\nJWT signing", x: 500, y: 55, w: 170, h: 75, color: colors.purple },
    { id: "db", label: "PostgreSQL\n:5432", x: 765, y: 55, w: 160, h: 76, kind: "db", color: colors.gray },
    { id: "mq", label: "RabbitMQ\n:5672 / :15672", x: 765, y: 225, w: 160, h: 86, kind: "db", color: colors.orange },
    { id: "notify", label: "Notification\nservice", x: 500, y: 410, w: 180, h: 80, color: colors.green },
    { id: "mailhog", label: "MailHog\nSMTP :1025\nUI :8025", x: 765, y: 405, w: 160, h: 92, color: colors.red },
  ],
  [
    { from: "browser", to: "frontend", label: "loads app" },
    { from: "browser", to: "api", label: "HTTP JSON + Bearer JWT" },
    { from: "api", to: "st", label: "role checks" },
    { from: "api", to: "db", label: "GORM" },
    { from: "api", to: "mq", label: "outbox events" },
    { from: "mq", to: "notify", label: "order facts" },
    { from: "notify", to: "mailhog", label: "SMTP" },
  ],
);

flow(
  "service-ownership",
  "Service Ownership",
  900,
  430,
  [
    { id: "ui", label: "Frontend\nmenu, cart, orders,\nstaff queue", x: 40, y: 160, w: 190, h: 100, color: colors.blue },
    { id: "auth", label: "Basic auth\nJWT middleware", x: 330, y: 40, w: 170, h: 80, color: colors.green },
    { id: "products", label: "Product/menu\nhandlers", x: 330, y: 155, w: 170, h: 80, color: colors.green },
    { id: "orders", label: "Order handlers\nworkflow", x: 330, y: 270, w: 170, h: 80, color: colors.green },
    { id: "outbox", label: "Transactional\noutbox", x: 580, y: 270, w: 170, h: 80, color: colors.yellow },
    { id: "consumer", label: "RabbitMQ\nconsumer", x: 580, y: 110, w: 170, h: 80, color: colors.orange },
    { id: "email", label: "Email formatting\nsender", x: 710, y: 210, w: 170, h: 80, color: colors.red },
  ],
  [
    { from: "ui", to: "auth" },
    { from: "ui", to: "products" },
    { from: "ui", to: "orders" },
    { from: "orders", to: "outbox" },
    { from: "outbox", to: "consumer" },
    { from: "consumer", to: "email" },
  ],
);

sequence("auth-role-sequence", "Auth And Role Flow", [
  { id: "user", label: "Browser\nuser" },
  { id: "ui", label: "React\nfrontend" },
  { id: "api", label: "Order\nservice" },
  { id: "headers", label: "Basic login\nJWT session" },
], [
  { from: "user", to: "ui", text: "Choose demo account" },
  { from: "ui", to: "headers", text: "POST /auth/login with Basic auth" },
  { from: "headers", to: "api", text: "Issue Bearer JWT" },
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
    { id: "staffq", label: "Role is\nstaff?", x: 700, y: 285, w: 120, h: 100, kind: "decision", color: colors.yellow },
    { id: "staff", label: "Role: staff", x: 740, y: 430, w: 140, h: 70, color: colors.orange },
    { id: "user", label: "Role:\ncustomer", x: 520, y: 475, w: 140, h: 70, color: colors.green },
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
  1050,
  620,
  [
    { id: "guest", label: "customer", x: 50, y: 55, w: 120, h: 60, color: colors.gray },
    { id: "user", label: "admin", x: 50, y: 195, w: 120, h: 60, color: colors.green },
    { id: "staff", label: "staff", x: 50, y: 340, w: 120, h: 60, color: colors.orange },
    { id: "admin", label: "admin", x: 50, y: 480, w: 120, h: 60, color: colors.purple },
    { id: "browse", label: "Browse\nproducts", x: 300, y: 40, w: 150, h: 70, color: colors.blue },
    { id: "create", label: "Create\norder", x: 510, y: 40, w: 150, h: 70, color: colors.blue },
    { id: "mine", label: "List own /\nemail orders", x: 720, y: 40, w: 170, h: 70, color: colors.blue },
    { id: "queue", label: "List all\norders", x: 300, y: 320, w: 150, h: 70, color: colors.yellow },
    { id: "status", label: "Ready, complete,\ncancel", x: 510, y: 320, w: 170, h: 70, color: colors.yellow },
    { id: "products", label: "Browse\nproducts", x: 720, y: 460, w: 150, h: 70, color: colors.red },
    { id: "delete", label: "Staff\nqueue", x: 900, y: 460, w: 120, h: 70, color: colors.red },
  ],
  [
    { from: "guest", to: "browse" },
    { from: "guest", to: "create" },
    { from: "guest", to: "mine" },
    { from: "user", to: "browse" },
    { from: "user", to: "create" },
    { from: "user", to: "mine" },
    { from: "staff", to: "queue" },
    { from: "staff", to: "status" },
    { from: "admin", to: "browse" },
    { from: "admin", to: "create" },
    { from: "admin", to: "mine" },
    { from: "admin", to: "queue" },
    { from: "admin", to: "status" },
    { from: "admin", to: "products" },
    { from: "admin", to: "delete" },
  ],
);

flow(
  "frontend-workflow",
  "Frontend Workflow",
  980,
  560,
  [
    { id: "start", label: "Open\nfrontend", x: 40, y: 240, w: 130, h: 70, color: colors.yellow },
    { id: "load", label: "Load\n/products", x: 230, y: 145, w: 130, h: 70, color: colors.blue },
    { id: "menu", label: "Menu\nview", x: 430, y: 145, w: 130, h: 70, color: colors.green },
    { id: "cart", label: "Add products\nto cart", x: 620, y: 145, w: 150, h: 70, color: colors.green },
    { id: "checkout", label: "Submit\n/orders", x: 815, y: 145, w: 130, h: 70, color: colors.orange },
    { id: "orders", label: "Orders\nview", x: 815, y: 300, w: 130, h: 70, color: colors.blue },
    { id: "mine", label: "Load\n/orders/mine", x: 620, y: 300, w: 150, h: 70, color: colors.blue },
    { id: "auth", label: "Login panel\n/demo creds", x: 230, y: 335, w: 150, h: 70, color: colors.purple },
    { id: "me", label: "POST /auth/login\nstore JWT", x: 430, y: 335, w: 130, h: 70, color: colors.purple },
    { id: "role", label: "Role?", x: 620, y: 420, w: 120, h: 90, kind: "decision", color: colors.yellow },
    { id: "staff", label: "Barista\nqueue", x: 815, y: 430, w: 130, h: 70, color: colors.orange },
  ],
  [
    { from: "start", to: "load" },
    { from: "load", to: "menu" },
    { from: "menu", to: "cart" },
    { from: "cart", to: "checkout" },
    { from: "checkout", to: "orders" },
    { from: "orders", to: "mine" },
    { from: "start", to: "auth" },
    { from: "auth", to: "me" },
    { from: "me", to: "role" },
    { from: "role", to: "menu", label: "customer/admin" },
    { from: "role", to: "staff", label: "staff/admin" },
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
