/*
THIS IS A GENERATED/BUNDLED FILE BY ESBUILD
if you want to view the source, please visit the github repository of this plugin
*/

var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// main.ts
var main_exports = {};
__export(main_exports, {
  default: () => AutoReadingMode
});
module.exports = __toCommonJS(main_exports);
var import_obsidian = require("obsidian");
var DEFAULT_SETTINGS = {
  timeout: 5,
  isReadingModeOnStartup: true
};
var AutoReadingMode = class extends import_obsidian.Plugin {
  constructor() {
    super(...arguments);
    this.timer = -1;
    this.shouldSetFirstMarkdownLeafToPreview = false;
  }
  async onload() {
    await this.loadSettings();
    this.app.workspace.onLayoutReady(() => {
      if (this.settings.isReadingModeOnStartup) {
        if (this.setMarkdownLeavesToPreviewMode() == 0) {
          this.shouldSetFirstMarkdownLeafToPreview = true;
        }
      }
    });
    this.registerEvent(
      this.app.workspace.on(
        "editor-change",
        this.resetTimeout.bind(this)
      )
    );
    this.registerEvent(
      this.app.workspace.on("active-leaf-change", (leaf) => {
        if (leaf == null)
          return;
        if (leaf.getViewState().type == "markdown" && this.shouldSetFirstMarkdownLeafToPreview) {
          this.setMarkdownLeavesToPreviewMode();
          this.shouldSetFirstMarkdownLeafToPreview = false;
        }
        this.resetTimeout();
      })
    );
    this.addSettingTab(new AutoReadingModeSettingTab(this.app, this));
  }
  onunload() {
    clearTimeout(this.timer);
  }
  async loadSettings() {
    this.settings = Object.assign(
      {},
      DEFAULT_SETTINGS,
      await this.loadData()
    );
  }
  async saveSettings() {
    await this.saveData(this.settings);
  }
  resetTimeout() {
    clearTimeout(this.timer);
    this.timer = window.setTimeout(() => {
      this.setMarkdownLeavesToPreviewMode();
    }, 6e4 * this.settings.timeout);
  }
  /**
   * Sets all active markdown leaves to preview mode.
   * @returns The number of markdown leaves present.
   */
  setMarkdownLeavesToPreviewMode() {
    const markdownLeaves = this.app.workspace.getLeavesOfType("markdown");
    markdownLeaves.forEach((workspaceLeaf) => {
      const viewState = workspaceLeaf.getViewState();
      workspaceLeaf.setViewState({
        ...viewState,
        state: { ...viewState.state, mode: "preview" }
      });
    });
    return markdownLeaves.length;
  }
};
var AutoReadingModeSettingTab = class extends import_obsidian.PluginSettingTab {
  constructor(app, plugin) {
    super(app, plugin);
    this.plugin = plugin;
  }
  display() {
    const { containerEl } = this;
    containerEl.empty();
    new import_obsidian.Setting(containerEl).setName("Timeout (minutes)").setDesc(
      "Timeout before Reading mode is enabled while Obsidian is active or minimized."
    ).addText(
      (text) => text.setValue(this.plugin.settings.timeout.toString()).onChange(async (value) => {
        const parsedInt = parseInt(value);
        if (Number.isNaN(parsedInt))
          return;
        this.plugin.settings.timeout = parsedInt;
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("Startup in Reading view").setDesc(
      "Show previously opened documents in Reading view when starting Obsidian."
    ).addToggle(
      (toggle) => toggle.setValue(this.plugin.settings.isReadingModeOnStartup).onChange(async (value) => {
        this.plugin.settings.isReadingModeOnStartup = value;
        await this.plugin.saveSettings();
      })
    );
  }
};
