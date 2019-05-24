const ColorOrder = { 'W': 0, 'U': 1, 'B': 2, 'R': 3, 'G': 4 }
const ColorShorts = { 'W': 'white', 'U': 'blue', 'B': 'black', 'R': 'red', 'G': 'green'}
const API = "api/v1/sessions"

const Languages = [
	{ code: 'en', name: 'English' },
	{ code: 'es', name: 'Spanish' },
	{ code: 'fr', name: 'French' },
	{ code: 'de', name: 'German' },
	{ code: 'it', name: 'Italian' },
	{ code: 'pt', name: 'Portuguese' },
	{ code: 'ja', name: 'Japanese' },
	{ code: 'ko', name: 'Korean' },
	{ code: 'ru', name: 'Russian' },
	{ code: 'zhs', name: 'Simplified Chinese' },
	{ code: 'zht', name: 'Traditional Chinese' }
]

window.onbeforeunload = function () {
    return "The session data is lost after the window is closed.";
}

const Sets = ["m19", "xln", "rix", "dom", "grn", "rna", "war"]

function orderColor(lhs, rhs) {
	if (!lhs || !rhs)
		return 0;
	if (lhs.length == 1 && rhs.length == 1)
		return ColorOrder[lhs[0]] - ColorOrder[rhs[0]];
	else if (lhs.length == 1)
		return -1;
	else if (rhs.length == 1)
		return 1;
	else
		return String(lhs.flat()).localeCompare(String(rhs.flat()));
}

Vue.component('card', {
	template: `
<figure class="card" :data-cmc="card.border_crop" v-on:click="action(card)">
	<img v-if="card.image_uris[language]" :src="card.image_uris[language]"/>
	<img v-else src="img/missing.svg">
	<figcaption>{{ card.count }}x {{ card.printed_name[language] }}</figcaption>
</figure>
`,
	props: ['card', 'language', 'action']
});

var app = new Vue({
	el: '#main-vue',
	data: {
		Session: sessionStorage.getItem("sessionid"),
		Player: null,
		SessionDetails: null,
		CardPool: null,
		Picks: null,
		Cards: null,
		Collection: null,
		
		// Session Creation Options
		Singleton: false,
		Set: "",
		ColorFilter: {
			white: true,
			blue: true,
			black: true,
			red: true,
			green: true,
			colorless: true,
		},
		RarityFilter: {
			common: true,
			uncommon: true,
			rare: true,
			mythic: true,
		},

		// View options
		Ready: false,
		Sets: Sets,
		CardOrder: "Color",
		DeckOrderCMC: true,
		HideCollectionManager: true,
		Languages: Languages,
		Language: 'en',
		ViewColor: {
			white: true,
			blue: true,
			black: true,
			red: true,
			green: true,
			colorless: true,
		},
		ViewCMC: {
			0: true,
			1: true,
			2: true,
			3: true,
			4: true,
			5: true,
			"6+": true,
		},

		// lobby socket
		websocket: null,
	},
	computed: {
		CardPoolUnsorted: function () {
			if (!this.CardPool || !this.Cards || !this.Picks) {
				return [];
			}

			let output = [];
			for (let cardID in this.CardPool) {
				details = this.getCard(cardID);
				details.count = this.CardPool[cardID];
				if (this.Picks[cardID]) {
					details.count -= this.Picks[cardID];
				}
				if (details.count > 0) {
					output.push(details);
				}
			}

			return output
		},
		CardPoolFiltered: function () {
			if (!this.CardPoolUnsorted) {
				return []
			}

			let colorFiltered = []
			for (let card of this.CardPoolUnsorted) {
				// special case colorless
				if (card.colors.length == 0) {
					if (this.ViewColor['colorless']) {
						colorFiltered.push(card)
					}
					continue
				}

				// main color filter
				for ( color of card.colors ) {
					if ( this.ViewColor[ColorShorts[color]] ) {
						colorFiltered.push(card)
						break
					}
				}
			}

			let cmcFiltered = []
			for (let card of colorFiltered) {
				// special case 6+
				if (card.cmc >= 6) {
					if (this.ViewCMC['6+']) {
						cmcFiltered.push(card)
					}
					continue
				}
				// main cmc filter
				if (this.ViewCMC[card.cmc]) {
					cmcFiltered.push(card)
				}
			}

			return cmcFiltered
		},
		CardPoolSorted: function () {
			if (this.CardOrder == 'CMC') {
				// sort by CMC, color, name (TODO name)
				return this.CardPoolFiltered.slice().sort(function (lhs, rhs) {
					if (lhs.cmc == rhs.cmc)
						return orderColor(lhs.colors, rhs.colors);
					return lhs.cmc > rhs.cmc;
				});
			}
			if (this.CardOrder == 'Color') {
				return this.CardPoolFiltered.slice().sort(function (lhs, rhs) {
					if (orderColor(lhs.colors, rhs.colors) == 0)
						return lhs.cmc > rhs.cmc;
					return orderColor(lhs.colors, rhs.colors);
				});
			}
		},
		DeckUnsorted: function () {
			if (!this.CardPool || !this.Picks || !this.Cards) {
				return [];
			}

			output = [];
			for (let cardID in this.Picks) {
				if (cardID == 0) {
					continue
				}
				details = this.getCard(cardID);
				for (let i = 0; i < this.Picks[cardID]; i++) {
					output.push(details);
				}
			}

			return output
		},
		DeckCMC: function () {
			let a = this.DeckUnsorted.reduce((acc, item) => {
				if (!acc[item.cmc])
					acc[item.cmc] = [];
				acc[item.cmc].push(item);
				return acc;
			}, {});
			return a;
		},
		SessionLobby: function () {
			if (!this.SessionDetails) {
				return null
			}

			let players = {};
			for (var key in this.SessionDetails['players']) {
				playerData = this.SessionDetails['players'][key]
				if (playerData['ready']) {
					players[key] = {'status': '✅',};
				} else {
					players[key] = {'status': '⛔',};
				}
				players[key].name = playerData['name']

				if (key == this.Player) {
					this.Ready = playerData.ready
				}
			}
			return players;
		},
		CollectionStats: function () {
			if (!this.Collection) {
				return 0;
			}

			let total = 0;
			for (let key in this.Collection) {
				total += this.Collection[key];
			}

			return total
		},
		CardPoolStats: function () {
			if (!this.CardPool) {
				return 0
			}

			let count = 0
			for (let card in this.CardPool) {
				count += this.CardPool[card]
			}
			return count;
		}
	},
	methods: {
		create_session() {
			if (!this.Collection) {
				return
			}

			fetch(API, {
				method: 'POST',
				body: JSON.stringify({
					name: name,
					collection: this.Collection,
					singleton: this.Singleton,
					rarity: this.RarityFilter,
					set: this.Set,
					color: this.ColorFilter,
				})
			}).then(function (response) {
				try {
					response.json().then(function (rep) {
						if (rep['id']) {
							app.set_session(rep['id']);
							return;
						}
						alert(rep['error']);
					});
				} catch (e) {
					alert(e);
				}
			});
		},
		textbox_set_session(event) {
			this.set_session(event.target.value);
		},
		set_session(s) {
			this.Session = s
			sessionStorage.setItem("sessionid", app.Session)
		},
		clear_session() {
			this.clear_registration();
			this.Session = null;
			sessionStorage.removeItem("sessionid")
			this.CardPool = null;
			sessionStorage.removeItem("cardpool")
			this.Picks = null;
			sessionStorage.removeItem("picks")
		},
		clear_registration() {
			this.Player = null;
			this.SessionDetails = null;

			if ( this.websocket ) {
				this.websocket.close()
				this.websocket = null
			}
		},
		textbox_register(event) {
			this.join_lobby(event.target.value);
		},
		join_lobby(name) {
			if ( name == "" ) {
				console.error("no name provided")
				return
			}

			if ( ! this.Session ) {
				return
			}

			if ( this.websocket ) {
				return
			}

			let ext = "/" + API + "/" + this.Session + "/players"

			this.websocket = new WebSocket(((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + ext );
			this.websocket.onopen = () => {
			  console.log("Connected to websocket")

			  // send player data
			  this.websocket.send(JSON.stringify({
				name: name,
				collection: this.Collection,
			  }))
			  
			  this.websocket.onmessage = ({data}) => {
				  try {
					let rep = JSON.parse(data)

					if (rep['error']) {
						alert("Error: " + rep['error'])
						app.clear_session()
						return
					}

					// first update should be a player registration
					if (!app.Player) {
						if (rep['id']) {
							app.Player = rep['id']
							return
						}
						alert("Error: no ID received")
						app.clear_registration()
						return
					}

					// subsequent updates: lobby updates
					app.SessionDetails = rep
					if (app.SessionDetails['started']) {
						app.load_card_pool()
					}
				  } catch(e) {
					  console.error(e)
				  }
			  };
			};
		},
		player_ready(event) {
			if (!this.Player || !this.Session || !this.websocket ) {
				return;
			}

			this.websocket.send(JSON.stringify({
				ready: event.target.checked,
			}))
		},
		load_card_pool() {
			if (!this.Session || !this.Player || !this.SessionDetails || !this.SessionDetails['started']) {
				return
			}

			fetch(API + "/" + this.Session + "/players/" + this.Player + "/collection").then(function (response) {
				try {
					response.json().then(function (rep) {
						if (rep['error']) {
							alert(rep['error']);
							return;
						}

						app.CardPool = rep
						sessionStorage.setItem("cardpool", JSON.stringify(rep))
						if (!app.Picks) {
							app.Picks = {}
							sessionStorage.setItem("picks", JSON.stringify(app.Picks))
						}

						// need to load card pool before closing websocket (session data is removed once all clients have disconnected)
						if ( app.websocket ) {
							app.websocket.close()
							app.websocket = null
						}
					});
				} catch (e) {
					alert(e);
				}
			});
		},
		pick(card) {
			if (!this.CardPool || !this.Picks) {
				return
			}

			let picked = 0;
			if (this.Picks[card.id]) {
				picked = this.Picks[card.id]
			}

			if (this.CardPool[card.id] - picked > 0) {
				this.$set(this.Picks, card.id, picked + 1);
				sessionStorage.setItem("picks", JSON.stringify(this.Picks))
			}
		},
		unpick(card) {
			if (!this.Picks) {
				return 
			}

			let picked = this.Picks[card.id];
			if (!picked) {
				return;
			}
			
			picked--;
			if (picked <= 0) {
				this.$delete(this.Picks, card.id);
			} else {
				this.$set(this.Picks, card.id, picked);
			}
			
			sessionStorage.setItem("picks", JSON.stringify(this.Picks))
		},
		parseMTGALog: function (e) {
			let file = e.target.files[0];
			if (!file) {
				return;
			}
			var reader = new FileReader();
			reader.onload = function (e) {
				let contents = e.target.result;
				let call_idx = contents.lastIndexOf("PlayerInventory.GetPlayerCardsV3");
				let collection_start = contents.indexOf('{', call_idx);
				let collection_end = contents.indexOf('}', collection_start);

				try {
					let collStr = contents.slice(collection_start, collection_end + 1);
					localStorage.setItem("Collection", collStr);
					let coll = JSON.parse(collStr)
					if (coll) {
						app.Collection = coll
					}
				} catch (e) {
					alert(e);
				}
			};
			reader.readAsText(file);
		},
		getCard(id) {
			if (!this.Cards || !this.Cards[id] || !id) {
				return
			}
			return {
				id: id,
				name: this.Cards[id].name,
				printed_name: this.Cards[id].printed_name,
				image_uris: this.Cards[id].image_uris,
				set: this.Cards[id].set,
				cmc: this.Cards[id].cmc,
				collector_number: this.Cards[id].collector_number,
				colors: this.Cards[id].color_identity,
			};
		}
	},
	created: function () {
		// Load card information
		fetch("data/MTGACards.json").then(function (response) {
			response.text().then(function (text) {
				try {
					// load cards
					tmpCards = JSON.parse(text)
					for (let c in app.Cards) {
						// populate all printed names and image uris if there is no resource for the given language
						for (let l of app.Languages) {
							if (!(l.code in app.Cards[c]['printed_name'])) {
								app.Cards[c]['printed_name'][l.code] = app.Cards[c]['name'];
							}
							if (!(l.code in app.Cards[c]['image_uris'])) {
								app.Cards[c]['image_uris'][l.code] = app.Cards[c]['image_uris']['en'];
							}
						}
					}
					app.Cards = tmpCards
				} catch (e) {
					alert(e);
				}
			});
		});

		// Look for a locally stored collection
		let localStorageCollection = localStorage.getItem("Collection")
		if (localStorageCollection) {
			try {
				this.Collection  = JSON.parse(localStorageCollection)
				console.log("Loaded collection from local storage")
			} catch (e) {
				console.error(e);
			}
		}

		// Look for locally stored picks
		let picks = sessionStorage.getItem("picks")
		if (picks) {
			try {
				this.Picks = JSON.parse(picks)
				console.log("Loaded picks from local storage")
			} catch (e) {
				console.error(e);
			}
		}

		// Look for locally stored cardpool
		let cardpool = sessionStorage.getItem("cardpool");
		if (cardpool) {
			try {
				this.CardPool = JSON.parse(cardpool);
				console.log("Loaded cardpool from local storage")
			} catch (e) {
				console.error(e);
			}
		}
	}
});

// Helper functions ////////////////////////////////////////////////////////////////////////////////

// https://hackernoon.com/copying-text-to-clipboard-with-javascript-df4d4988697f
const copyToClipboard = str => {
	const el = document.createElement('textarea');  // Create a <textarea> element
	el.value = str;                                 // Set its value to the string that you want copied
	el.setAttribute('readonly', '');                // Make it readonly to be tamper-proof
	el.style.position = 'absolute';
	el.style.left = '-9999px';                      // Move outside the screen to make it invisible
	document.body.appendChild(el);                  // Append the <textarea> element to the HTML document
	const selected =
		document.getSelection().rangeCount > 0      // Check if there is any content selected previously
			? document.getSelection().getRangeAt(0) // Store selection if found
			: false;                                // Mark as false to know no selection existed before
	el.select();                                    // Select the <textarea> content
	document.execCommand('copy');                   // Copy - only works as a result of a user action (e.g. click events)
	document.body.removeChild(el);                  // Remove the <textarea> element
	if (selected) {                                 // If a selection existed before copying
		document.getSelection().removeAllRanges();  // Unselect everything on the HTML document
		document.getSelection().addRange(selected); // Restore the original selection
	}
};

function exportMTGA(deckUnsorted) {
	let str = "";
	for (card of deckUnsorted) {
		let set = card.set.toUpperCase();
		if (set == "DOM") set = "DAR"; // DOM is called DAR in MTGA
		let name = card.printed_name[app.Language];

		// multi-card handling
		let idx = name.indexOf('//');
		if (idx != -1) {
			name = name.substr(0, idx - 1)
		}

		str += `1 ${name} (${set}) ${card.collector_number}\n`
	}
	return str;
}