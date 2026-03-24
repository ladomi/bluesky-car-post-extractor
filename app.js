const dom = {
  dropzone: document.querySelector("#dropzone"),
  fileInput: document.querySelector("#car-file"),
  selectedFile: document.querySelector("#selected-file"),
  extractButton: document.querySelector("#extract-button"),
  runtimeStatus: document.querySelector("#runtime-status"),
  resultsPanel: document.querySelector("#results-panel"),
  didValue: document.querySelector("#did-value"),
  postCount: document.querySelector("#post-count"),
  rangeStart: document.querySelector("#range-start"),
  rangeEnd: document.querySelector("#range-end"),
  downloadJsonLink: document.querySelector("#download-json-link"),
  downloadCsvLink: document.querySelector("#download-csv-link"),
  downloadName: document.querySelector("#download-name"),
  previewList: document.querySelector("#preview-list"),
};

let currentFile = null;
let currentJsonDownloadUrl = null;
let currentCsvDownloadUrl = null;

bootExtractor();
bindUi();

async function bootExtractor() {
  try {
    const go = new Go();
    const response = await fetch("./car-extractor.wasm");
    if (!response.ok) {
      throw new Error(`failed to fetch wasm: ${response.status}`);
    }

    let instance;
    try {
      ({ instance } = await WebAssembly.instantiateStreaming(response.clone(), go.importObject));
    } catch {
      ({ instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject));
    }

    go.run(instance);
    await waitForExtractor();
    setRuntimeStatus("準備完了", "ready");
    dom.extractButton.disabled = !currentFile;
  } catch (error) {
    console.error(error);
    setRuntimeStatus("WASM の読み込みに失敗", "error");
  }
}

function bindUi() {
  dom.fileInput.addEventListener("change", () => {
    updateCurrentFile(dom.fileInput.files?.[0] ?? null);
  });

  dom.extractButton.addEventListener("click", async () => {
    if (!currentFile) {
      return;
    }
    await extractCurrentFile();
  });

  dom.dropzone.addEventListener("dragover", (event) => {
    event.preventDefault();
    dom.dropzone.dataset.dragging = "true";
  });

  dom.dropzone.addEventListener("dragleave", () => {
    delete dom.dropzone.dataset.dragging;
  });

  dom.dropzone.addEventListener("drop", (event) => {
    event.preventDefault();
    delete dom.dropzone.dataset.dragging;

    const file = event.dataTransfer?.files?.[0] ?? null;
    if (!file) {
      return;
    }

    const transfer = new DataTransfer();
    transfer.items.add(file);
    dom.fileInput.files = transfer.files;
    updateCurrentFile(file);
  });
}

function updateCurrentFile(file) {
  currentFile = file;
  if (!file) {
    dom.selectedFile.textContent = "未選択";
    dom.extractButton.disabled = true;
    return;
  }

  dom.selectedFile.textContent = `${file.name} (${formatBytes(file.size)})`;
  dom.extractButton.disabled = dom.runtimeStatus.dataset.state !== "ready";
}

async function extractCurrentFile() {
  if (typeof globalThis.extractBlueskyPostsFromCar !== "function") {
    setRuntimeStatus("WASM 初期化が完了していません", "error");
    return;
  }

  setRuntimeStatus("CAR を解析中...", "busy");
  dom.extractButton.disabled = true;
  dom.resultsPanel.hidden = true;

  try {
    const bytes = new Uint8Array(await currentFile.arrayBuffer());
    await nextPaint();

    const raw = globalThis.extractBlueskyPostsFromCar(bytes);
    const result = JSON.parse(raw);

    if (result.error) {
      throw new Error(result.error);
    }

    renderResult(result);
    setRuntimeStatus("抽出完了", "ready");
  } catch (error) {
    console.error(error);
    setRuntimeStatus("抽出に失敗", "error");
    dom.previewList.replaceChildren(renderMessage(error.message));
    dom.resultsPanel.hidden = false;
  } finally {
    dom.extractButton.disabled = dom.runtimeStatus.dataset.state !== "ready" || !currentFile;
  }
}

function renderResult(result) {
  dom.didValue.textContent = result.did || "-";
  dom.postCount.textContent = String(result.postCount ?? 0);
  dom.rangeStart.textContent = result.rangeStart || "-";
  dom.rangeEnd.textContent = result.rangeEnd || "-";

  const jsonDownloadName = buildDownloadName(currentFile?.name ?? "posts.car", "json");
  const csvDownloadName = buildDownloadName(currentFile?.name ?? "posts.car", "csv");
  setDownloadLink("json", jsonDownloadName, result.downloadJson ?? "[]");
  setDownloadLink("csv", csvDownloadName, result.downloadCsv ?? "");
  dom.downloadName.textContent = `JSON: ${jsonDownloadName} / CSV: ${csvDownloadName}`;

  const fragment = document.createDocumentFragment();
  const preview = Array.isArray(result.preview) ? result.preview : [];

  if (preview.length === 0) {
    fragment.append(renderMessage("投稿レコードは見つかりませんでした。"));
  } else {
    for (const post of preview) {
      const article = document.createElement("article");
      article.className = "preview-card";

      const time = document.createElement("time");
      time.className = "preview-time";
      time.dateTime = post.createdAt;
      time.textContent = post.createdAt;

      const text = document.createElement("p");
      text.className = "preview-text";
      text.textContent = post.text || "(empty post text)";

      article.append(time, text);
      fragment.append(article);
    }
  }

  dom.previewList.replaceChildren(fragment);
  dom.resultsPanel.hidden = false;
}

function setDownloadLink(format, filename, bodyText) {
  if (format === "json" && currentJsonDownloadUrl) {
    URL.revokeObjectURL(currentJsonDownloadUrl);
  }

  if (format === "csv" && currentCsvDownloadUrl) {
    URL.revokeObjectURL(currentCsvDownloadUrl);
  }

  const mimeType =
    format === "csv" ? "text/csv;charset=utf-8" : "application/json;charset=utf-8";
  const downloadUrl = URL.createObjectURL(new Blob([bodyText], { type: mimeType }));

  if (format === "csv") {
    currentCsvDownloadUrl = downloadUrl;
    dom.downloadCsvLink.href = downloadUrl;
    dom.downloadCsvLink.download = filename;
    return;
  }

  currentJsonDownloadUrl = downloadUrl;
  dom.downloadJsonLink.href = downloadUrl;
  dom.downloadJsonLink.download = filename;
}

function buildDownloadName(fileName, extension) {
  const baseName = fileName.replace(/\.car$/i, "") || "bluesky-posts";
  return `${baseName}.posts.${extension}`;
}

function setRuntimeStatus(message, state) {
  dom.runtimeStatus.textContent = message;
  dom.runtimeStatus.dataset.state = state;
}

function renderMessage(message) {
  const article = document.createElement("article");
  article.className = "preview-card preview-card-message";
  article.textContent = message;
  return article;
}

function formatBytes(bytes) {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function nextPaint() {
  return new Promise((resolve) => requestAnimationFrame(() => resolve()));
}

async function waitForExtractor() {
  for (let index = 0; index < 200; index += 1) {
    if (typeof globalThis.extractBlueskyPostsFromCar === "function") {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 25));
  }
  throw new Error("extractor did not initialize");
}
