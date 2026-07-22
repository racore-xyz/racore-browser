import { a as require_react, s as __toESM, t as require_jsx_runtime } from "../index.js";
//#region app/page.tsx
var import_react = /* @__PURE__ */ __toESM(require_react(), 1);
var import_jsx_runtime = require_jsx_runtime();
var providers = [
	"OpenAI",
	"Anthropic",
	"Gemini",
	"OpenRouter",
	"Kimi",
	"Ollama",
	"OpenCode",
	"Claude Code",
	"Kimi Code"
];
var architecture = [
	{
		n: "01",
		title: "Agentic browser",
		body: "A Chromium desktop workspace where AI can research, navigate, and act—while every meaningful side effect waits for you."
	},
	{
		n: "02",
		title: "Local control plane",
		body: "The Go daemon keeps provider keys, approvals, identity, publishing, and peer connections on your machine."
	},
	{
		n: "03",
		title: "Open web protocol",
		body: "Signed releases, portable domains, content-addressed storage, and a peer mesh make sites verifiable beyond one platform."
	}
];
var flow = [
	[
		"01",
		"ASK",
		"Give Racore a goal in natural language."
	],
	[
		"02",
		"PLAN",
		"The agent chooses tools and the right provider."
	],
	[
		"03",
		"APPROVE",
		"You confirm actions that change the outside world."
	],
	[
		"04",
		"VERIFY",
		"Racore returns evidence and signed outputs."
	]
];
function Arrow() {
	return /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
		"aria-hidden": "true",
		children: "↗"
	});
}
function Home() {
	const [menuOpen, setMenuOpen] = (0, import_react.useState)(false);
	const [progress, setProgress] = (0, import_react.useState)(0);
	(0, import_react.useEffect)(() => {
		const updateProgress = () => {
			const max = document.documentElement.scrollHeight - window.innerHeight;
			setProgress(max > 0 ? window.scrollY / max : 0);
		};
		updateProgress();
		window.addEventListener("scroll", updateProgress, { passive: true });
		const observer = new IntersectionObserver((entries) => entries.forEach((entry) => entry.isIntersecting && entry.target.classList.add("is-visible")), { threshold: .12 });
		document.querySelectorAll("[data-reveal]").forEach((element) => observer.observe(element));
		return () => {
			window.removeEventListener("scroll", updateProgress);
			observer.disconnect();
		};
	}, []);
	return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("main", {
		className: "landing",
		onPointerMove: (event) => {
			const x = event.clientX / window.innerWidth - .5;
			const y = event.clientY / window.innerHeight - .5;
			event.currentTarget.style.setProperty("--pointer-x", `${x * 10}deg`);
			event.currentTarget.style.setProperty("--pointer-y", `${y * -8}deg`);
		},
		children: [
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
				className: "scroll-progress",
				style: { transform: `scaleX(${progress})` }
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("header", {
				className: "landing-nav",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
						className: "landing-logo",
						href: "#top",
						"aria-label": "Racore home",
						children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)("img", {
							src: "/brand/racore-logo.png",
							alt: "Racore.xyz",
							width: "202",
							height: "52"
						})
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("nav", {
						className: menuOpen ? "nav-links is-open" : "nav-links",
						"aria-label": "Main navigation",
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
								href: "#system",
								onClick: () => setMenuOpen(false),
								children: "System"
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
								href: "#protocol",
								onClick: () => setMenuOpen(false),
								children: "Protocol"
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
								href: "#providers",
								onClick: () => setMenuOpen(false),
								children: "Providers"
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
								href: "#safety",
								onClick: () => setMenuOpen(false),
								children: "Safety"
							})
						]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("a", {
						className: "nav-launch",
						href: "/browser",
						children: ["Launch browser ", /* @__PURE__ */ (0, import_jsx_runtime.jsx)(Arrow, {})]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("button", {
						className: "menu-toggle",
						onClick: () => setMenuOpen((open) => !open),
						"aria-expanded": menuOpen,
						"aria-label": "Toggle navigation",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {})]
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
				className: "hero",
				id: "top",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "hero-meta",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "RACORE / 0.1" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "AGENTIC BROWSER + OPEN WEB PROTOCOL" })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "hero-visual",
						"aria-hidden": "true",
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
								className: "hero-orbit",
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {})
								]
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("img", {
								className: "hero-core-image",
								src: "/generated/racore-core.png",
								alt: ""
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
								className: "hero-image-label",
								children: "RACORE INTELLIGENCE CORE / LIVE"
							})
						]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "hero-copy",
						"data-reveal": true,
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
								className: "eyebrow",
								children: "Browse with agents. Publish without lock-in."
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("h1", { children: [
								"The browser",
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
								"built for ",
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("em", { children: "agency." })
							] }),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
								className: "hero-bottom",
								children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "Racore brings AI, identity, publishing, and a peer-to-peer web into one local-first workspace—without surrendering control." }), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
									className: "hero-actions",
									children: [/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("a", {
										className: "button button-dark",
										href: "/browser",
										children: ["Enter Racore ", /* @__PURE__ */ (0, import_jsx_runtime.jsx)(Arrow, {})]
									}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("a", {
										className: "text-link",
										href: "#system",
										children: ["Explore the system ", /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "↓" })]
									})]
								})]
							})
						]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "hero-status",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "SCROLL TO DISCOVER" }), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("span", {
							className: "status-live",
							children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}), " LOCAL-FIRST / ONLINE"]
						})]
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
				className: "statement section-grid",
				id: "system",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "section-kicker",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "01" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "THE RACORE SYSTEM" })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "statement-copy",
						"data-reveal": true,
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
							className: "eyebrow",
							children: "One interface. Three layers."
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("h2", { children: [
							"From browsing the web",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
							"to ",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("em", { children: "owning" }),
							" the outcome."
						] })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
						className: "architecture-grid",
						children: architecture.map((item) => /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", {
							"data-reveal": true,
							children: [
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: item.n }),
								/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
									className: "architecture-icon",
									"aria-hidden": "true",
									children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {})]
								}),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h3", { children: item.title }),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: item.body }),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
									href: item.n === "01" ? "/browser" : "#protocol",
									"aria-label": `Learn about ${item.title}`,
									children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)(Arrow, {})
								})
							]
						}, item.n))
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
				className: "network-section",
				id: "protocol",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "network-top",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "02 / OPEN WEB INFRASTRUCTURE" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "(LIVE SYSTEM MODEL)" })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "network-title",
						"data-reveal": true,
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
							className: "eyebrow",
							children: "Your work should outlive a platform"
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("h2", { children: [
							"A web with",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("em", { children: "memory & proof." })
						] })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "network-stage",
						"aria-label": "Racore decentralized network illustration",
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
								className: "network-rings",
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {})
								]
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
								className: "network-core-wrap",
								children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)("img", {
									className: "network-core-image",
									src: "/generated/racore-core.png",
									alt: ""
								})
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
								className: "node node-a",
								children: "SIGNED RELEASES"
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
								className: "node node-b",
								children: "IPFS / KUBO"
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
								className: "node node-c",
								children: "DOMAIN AUTHORITY"
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
								className: "node node-d",
								children: "RACOON MESH"
							})
						]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "network-foot",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "Every release can be content-addressed, signed by its owner, and shared across a discovery mesh. Verification travels with the work." }), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("a", {
							className: "button button-light",
							href: "#architecture",
							children: ["See architecture ", /* @__PURE__ */ (0, import_jsx_runtime.jsx)(Arrow, {})]
						})]
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
				className: "metrics section-grid",
				id: "safety",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "section-kicker",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "03" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "CONTROL BY DESIGN" })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "metrics-heading",
						"data-reveal": true,
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
							className: "eyebrow",
							children: "Automation with boundaries"
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("h2", { children: [
							"AI that knows",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
							"when to ",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("em", { children: "ask." })
						] })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "metric-grid",
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", {
								"data-reveal": true,
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "01" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("strong", { children: ["100", /* @__PURE__ */ (0, import_jsx_runtime.jsx)("sup", { children: "%" })] }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h3", { children: "Explicit approval" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "External side effects stop at a human checkpoint before execution." })
								]
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", {
								"data-reveal": true,
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "02" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("strong", { children: "127.0.0.1" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h3", { children: "Local control plane" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "The daemon exposes its API on the loopback interface, close to your data." })
								]
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", {
								"data-reveal": true,
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "03" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("strong", { children: ["AES", /* @__PURE__ */ (0, import_jsx_runtime.jsx)("sup", { children: "256" })] }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h3", { children: "Encrypted secrets" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "Provider credentials are protected at rest in the local vault." })
								]
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", {
								"data-reveal": true,
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "04" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("strong", { children: "0" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h3", { children: "Policy bypasses" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "No CAPTCHA evasion, access-control bypass, rate-limit avoidance, or ban circumvention." })
								]
							})
						]
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
				className: "providers section-grid",
				id: "providers",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "section-kicker",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "04" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "PROVIDER FREEDOM" })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "provider-heading",
						"data-reveal": true,
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
								className: "eyebrow",
								children: "Bring the intelligence you trust"
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("h2", { children: [
								"Nine providers.",
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
								"One ",
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("em", { children: "gateway." })
							] }),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "Connect cloud models, local runtimes, or coding agents. Racore routes requests through a consistent interface while keys remain under your control." })
						]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
						className: "provider-stack",
						"aria-label": "Supported AI providers",
						children: providers.map((provider, index) => /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
							style: { "--i": index },
							children: [
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: String(index + 1).padStart(2, "0") }),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("b", { children: provider }),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", { children: index < 5 ? "CLOUD" : "LOCAL" })
							]
						}, provider))
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
				className: "workflow section-grid",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "section-kicker",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "05" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "HOW IT MOVES" })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "workflow-title",
						"data-reveal": true,
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
							className: "eyebrow",
							children: "A visible chain of intent"
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("h2", { children: [
							"From a goal",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
							"to a ",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("em", { children: "verified result." })
						] })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
						className: "flow-grid",
						children: flow.map(([n, title, body]) => /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", {
							"data-reveal": true,
							children: [
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: n }),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
									className: "flow-glyph",
									children: title.slice(0, 1)
								}),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h3", { children: title }),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: body })
							]
						}, n))
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
				className: "architecture-map",
				id: "architecture",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "map-header",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "06 / REPOSITORY ARCHITECTURE" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "WINDOWS · WEB · PEER NETWORK" })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "map-title",
						"data-reveal": true,
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
							className: "eyebrow",
							children: "Built as interoperable layers"
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("h2", { children: [
							"One product.",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
							"Clear ",
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("em", { children: "responsibilities." })
						] })]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "system-map",
						"data-reveal": true,
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
								className: "map-column",
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "INTERFACE" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", { children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("b", { children: "React + Next.js" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "Agentic browser, sites, providers, network, and system views." })] }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", { children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("b", { children: "Electron / Tauri" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "Desktop packaging and a secure bridge to the local runtime." })] })
								]
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
								className: "map-spine",
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {}),
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("i", {})
								]
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
								className: "map-column map-column-right",
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "CONTROL PLANE" }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", { children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("b", { children: "racored / Go" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "REST + WebSocket API, approvals, gateway, vault, and orchestration." })] }),
									/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("article", { children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("b", { children: "Protocol + storage" }), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", { children: "DID identity, signed messages, domain authority, mesh, and IPFS." })] })
								]
							})
						]
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
				className: "final-cta",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
						className: "cta-noise",
						"aria-hidden": "true"
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						"data-reveal": true,
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
								className: "eyebrow",
								children: "The web can be useful and yours"
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("h2", { children: [
								"Browse boldly.",
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
								/* @__PURE__ */ (0, import_jsx_runtime.jsx)("em", { children: "Keep control." })
							] }),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("a", {
								className: "button button-light",
								href: "/browser",
								children: ["Launch Racore ", /* @__PURE__ */ (0, import_jsx_runtime.jsx)(Arrow, {})]
							})
						]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
						className: "cta-code",
						children: "R / 2026 / OPEN AGENTIC WEB"
					})
				]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("footer", {
				className: "landing-footer",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
						className: "footer-logo",
						href: "#top",
						children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)("img", {
							src: "/brand/racore-logo.png",
							alt: "Racore.xyz",
							width: "190",
							height: "48"
						})
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("p", { children: [
						"Agentic browser & open web protocol.",
						/* @__PURE__ */ (0, import_jsx_runtime.jsx)("br", {}),
						"Designed for human agency."
					] }),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", { children: [
						/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
							href: "#system",
							children: "System"
						}),
						/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
							href: "#protocol",
							children: "Protocol"
						}),
						/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
							href: "#providers",
							children: "Providers"
						}),
						/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
							href: "/browser",
							children: "Launch"
						})
					] }),
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", { children: "© 2026 RACORE.XYZ" })
				]
			})
		]
	});
}
//#endregion
export { Home as default };
