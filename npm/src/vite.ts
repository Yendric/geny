import type { Plugin } from "vite";

const runtimeEntry = "virtual:geny/islands";
const resolvedRuntimeEntry = "\0" + runtimeEntry;

const islandExtensions = ["tsx", "jsx"];

function islandsDir(): string {
  const dir = process.env.GENY_ISLANDS_DIR ?? "islands";
  return "/" + dir.replace(/^\.?\/+|\/+$/g, "") + "/";
}

export default function geny(): Plugin {
  let command: "build" | "serve";
  return {
    name: "geny",
    config(_, env) {
      command = env.command;
      return {
        resolve: { dedupe: ["react", "react-dom"] },
        optimizeDeps: { include: ["react", "react-dom/client"] },
        server: { watch: { ignored: ["**/.geny/**"] } },
      };
    },
    buildStart() {
      if (command === "build") {
        this.emitFile({ type: "chunk", id: runtimeEntry });
      }
    },
    resolveId(id) {
      return id === runtimeEntry ? resolvedRuntimeEntry : undefined;
    },
    load(id) {
      if (id !== resolvedRuntimeEntry) {
        return undefined;
      }
      const dir = islandsDir();
      const patterns = islandExtensions.map((ext) => `${dir}**/*.${ext}`);
      return `import { start } from "@yendric/geny/runtime";
start(import.meta.glob(${JSON.stringify(patterns)}), ${JSON.stringify(dir)});`;
    },
  };
}
