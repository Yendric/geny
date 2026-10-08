import { existsSync } from "node:fs";
import { resolve, sep } from "node:path";
import type { Plugin } from "vite";

const runtimeEntry = "virtual:geny/islands";
const resolvedRuntimeEntry = "\0" + runtimeEntry;

const islandExtensions = ["tsx", "jsx"];

function dirFromEnv(name: string, fallback: string): string {
  return (process.env[name] ?? fallback).replace(/^\.?\/+|\/+$/g, "");
}

export default function geny(): Plugin {
  const islandsDir = dirFromEnv("GENY_ISLANDS_DIR", "islands");
  const reloadDirs = [
    dirFromEnv("GENY_CONTENT_DIR", "content"),
    dirFromEnv("GENY_TEMPLATES_DIR", "templates"),
    dirFromEnv("GENY_PUBLIC_DIR", "public"),
  ];
  let command: "build" | "serve";
  let hasIslands = false;

  return {
    name: "geny",
    config(config, env) {
      command = env.command;
      hasIslands = existsSync(resolve(config.root ?? process.cwd(), islandsDir));
      const server = { watch: { ignored: ["**/.geny/**"] } };
      if (!hasIslands) {
        return { server };
      }
      return {
        server,
        resolve: { dedupe: ["react", "react-dom"] },
        optimizeDeps: { include: ["react", "react-dom/client"] },
      };
    },
    // geny holds page requests until its rebuild finishes, so reloading on the source change is safe
    configureServer(server) {
      const dirs = reloadDirs.map((dir) => resolve(server.config.root, dir));
      server.watcher.add(dirs);
      server.watcher.on("all", (_event, file) => {
        if (dirs.some((dir) => file === dir || file.startsWith(dir + sep))) {
          server.ws.send({ type: "full-reload" });
        }
      });
    },
    buildStart() {
      if (command === "build" && hasIslands) {
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
      const dir = `/${islandsDir}/`;
      const patterns = islandExtensions.map((ext) => `${dir}**/*.${ext}`);
      return `import { start } from "@yendric/geny/runtime";
start(import.meta.glob(${JSON.stringify(patterns)}), ${JSON.stringify(dir)});`;
    },
  };
}
