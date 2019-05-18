const ColorOrder = { 'W': 0, 'U': 1, 'B': 2, 'R': 3, 'G': 4 }
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
		Session: localStorage.getItem("sessionid"),
		Player: localStorage.getItem("playerid"),
		SessionDetails: null,
		Ready: false,
		CardPool: {},
		Picks: localStorage.getItem("picks"),
		Cards: null,
		Collection: null,
		CollectionDate: "",
		
		// Session Creation Options
		Singleton: false,
		Pauper: false,
		Set: "",
		// BoosterQuantity: 6,
		// SetRestriction: "",

		// View options
		Sets: Sets,
		CardOrder: "Color",
		DeckOrderCMC: true,
		HideCollectionManager: true,
		Languages: Languages,
		Language: 'en',
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

			return output;
		},
		CardPoolCMC: function () {
			// sort by CMC, color, name (TODO name)
			return this.CardPoolUnsorted.slice().sort(function (lhs, rhs) {
				if (lhs.cmc == rhs.cmc)
					return orderColor(lhs.colors, rhs.colors);
				return lhs.cmc > rhs.cmc;
			});
		},
		CardPoolColor: function () {
			// sort by color, cmc, name (TODO name)
			return this.CardPoolUnsorted.slice().sort(function (lhs, rhs) {
				if (orderColor(lhs.colors, rhs.colors) == 0)
					return lhs.cmc > rhs.cmc;
				return orderColor(lhs.colors, rhs.colors);
			});
		},
		DeckUnsorted: function () {
			if (!this.CardPool || !this.Picks || !this.Cards) {
				return [];
			}

			output = [];
			for (let cardID in this.Picks) {
				details = this.getCard(cardID);
				for (let i = 0; i<this.Picks[cardID];i++) {
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
			if ( ! this.SessionDetails ) {
				return {}
			}

			let players = {};
			for (var key in this.SessionDetails['players']) {
				if (this.SessionDetails['players'][key]['ready']) {
					players[this.SessionDetails['players'][key]['name']] = '✅';
				} else {
					players[this.SessionDetails['players'][key]['name']] = '⛔️';
				}

				if (key == this.Player) {
					this.Ready = this.SessionDetails['players'][key]['ready'];
				}
			}
			return players;
		},
		Started: function () {
			if (!this.SessionDetails || !this.SessionDetails['started']) {
				return false
			}

			return this.SessionDetails['started']
		},
		CollectionStats: function () {
			if ( ! this.Collection) {
				return 0;
			}

			let total = 0;
			for (let key in this.Collection) {
				total += this.Collection[key];
			}

			return total
		},
		CardPoolStats: function () {
			if ( ! this.CardPool ) {
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
			if ( ! this.Collection ) {
				return
			}

			fetch(API, {
				method: 'POST',
				body: JSON.stringify({
					name: name,
					collection: this.Collection,
					singleton: this.Singleton,
					pauper: this.Pauper,
					set: this.Set,
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
		set_session(s) {
			this.Session = s
			this.verify_session()
		},
		textbox_set_session(event) {
			this.set_session(event.target.value);
		},
		verify_session() {
			if (!this.Session) {
				this.clear_session()
				return;
			}
			fetch(API + "/" + this.Session).then(function (response) {
				try {
					response.json().then(function (rep) {
						if (rep['error']) {
							console.log("session not found: " + app.Session);
							app.clear_session();
							return;
						}

						localStorage.setItem("sessionid", app.Session)
					});
				} catch (e) {
					alert(e);
				}
			});
		},
		verify_player() {
			if (!this.Session) {
				this.clear_session();
				return;
			}
			if (!this.Player) {
				this.clear_registration();
				return;
			}

			fetch(API + "/" + this.Session).then(function (response) {
				try {
					response.json().then(function (rep) {
						if (rep['error']) {
							console.log("session not found: " + app.Session);
							app.clear_session();
							return;
						}
						if (!rep['players'][app.Player]) {
							console.log("player not found");
							app.clear_registration();
							return;
						}

						localStorage.setItem("playerid", app.Player)
						app.set_session_details(rep)
					});
				} catch (e) {
					alert(e);
				}
			});
		},
		clear_session() {
			console.log("clearing session");
			this.clear_registration();
			this.Session = null;
			localStorage.removeItem("sessionid")
			this.CardPool = null;
			this.Picks = null;
			localStorage.removeItem("picks")
			//this.Boosters = [];
		},
		clear_registration() {
			this.Player = null;
			this.SessionDetails = null;
			localStorage.removeItem("playerid")
		},
		textbox_register(event) {
			this.register_player(event.target.value);
		},
		register_player(name) {
			if ( ! this.Session || ! this.Collection ) {
				return;
			}
			fetch(API + "/" + this.Session + "/players", {
				method: 'POST',
				body: JSON.stringify({
					name: name,
					collection: this.Collection,
				})
			}).then(function (response) {
				try {
					response.json().then(function (rep) {
						if (rep['id']) {
							app.set_player(rep['id']);
							return;
						}
						alert(rep['error']);
						clear_registration();
					});
				} catch (e) {
					alert(e);
				}
			});
		},
		set_player(id) {
			this.Player = id
			this.verify_player()
		},
		player_ready(event) {
			if (!this.Player || !this.Session) {
				return;
			}

			fetch(API + "/" + this.Session + "/players/" + this.Player, {
				method: 'POST',
				body: JSON.stringify({
					ready: event.target.checked,
				})
			}).then(function (response) {
				try {
					response.json().then(function (rep) {
						if (rep['error']) {
							alert(rep['error']);
							return;
						}

						app.set_session_details(rep);
					});
				} catch (e) {
					alert(e);
				}
			});
		},
		refresh_session() {
			if (!this.Session) {
				return
			}

			fetch(API + "/" + this.Session).then(function (response) {
				try {
					response.json().then(function (rep) {
						app.set_session_details(rep);
					});
				} catch {
					alert(e)
				}
			});
		},
		load_card_pool() {
			if (!this.Session || !this.Player || !this.Started) {
				return
			}

			fetch(API + "/" + this.Session + "/players/" + this.Player + "/collection").then(function (response) {
				try {
					response.json().then(function (rep) {
						if (rep['error']) {
							alert(rep['error']);
							return;
						}

						app.CardPool = rep;
						app.Picks = {};
					});
				} catch (e) {
					alert(e);
				}
			});
		},
		set_session_details(details) {
			var started = (this.SessionDetails && this.SessionDetails['started']);
			this.SessionDetails = details;

			// started now, not started before
			if (details['started'] && !started) {
				this.load_card_pool();
			}
		},
		pick(card) {
			if ( ! this.CardPool || ! this.Picks) {
				return
			}

			let picked = 0;
			if (this.Picks[card.id]) {
				picked = this.Picks[card.id]
			}

			if (this.CardPool[card.id] - picked > 0) {
				this.$set(this.Picks, card.id, picked + 1);
				localStorage.setItem("picks", JSON.stringify(this.Picks))
			}
		},
		unpick(card) {
			if ( ! this.Picks ) {
				return 
			}

			let picked = this.Picks[card.id];
			if (! picked) {
				return;
			}
			
			picked--;
			if (picked <= 0) {
				this.$delete(this.Picks, card.id);
				localStorage.setItem("picks", JSON.stringify(this.Picks))
				return;
			}

			this.$set(this.Picks, card.id, picked);
			localStorage.setItem("picks", JSON.stringify(this.Picks))
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
					if ( coll ) {
						app.Collection = coll
					}
				} catch (e) {
					alert(e);
				}
			};
			reader.readAsText(file);
		},
		getCard(id) {
			return {
				id: id,
				name: app.Cards[id].name,
				printed_name: app.Cards[id].printed_name,
				image_uris: app.Cards[id].image_uris,
				set: app.Cards[id].set,
				cmc: app.Cards[id].cmc,
				collector_number: app.Cards[id].collector_number,
				colors: app.Cards[id].color_identity,
			};
		}
	},
	mounted: function () {
		this.verify_session();
		this.verify_player()
		// Load all card informations
		fetch("data/MTGACards.json").then(function (response) {
			response.text().then(function (text) {
				try {
					tmpCards = JSON.parse(text);
					for (let c in app.Cards) {
						// populate all printed names and image uris if there is no resource for the given language
						for (let l of app.Languages) {
							if (!(l.code in app.Cards[c]['printed_name']))
								app.Cards[c]['printed_name'][l.code] = app.Cards[c]['name'];
							if (!(l.code in app.Cards[c]['image_uris']))
								app.Cards[c]['image_uris'][l.code] = app.Cards[c]['image_uris']['en'];
						}
					}
					app.Cards = tmpCards

					// Look for a localy stored collection
					let localStorageCollection = localStorage.getItem("Collection");
					if (localStorageCollection) {
						try {
							let json = JSON.parse(localStorageCollection);
							if ( json ) {
								app.Collection = json
							}
							console.log("Loaded collection from local storage");
						} catch (e) {
							console.error(e);
						}
					}
				} catch (e) {
					alert(e);
				}
			});
		});
	}
});

// Helper functions ////////////////////////////////////////////////////////////////////////////////

function isEmpty(obj) {
	console.log()
	for (var key in obj) {
		if (obj.hasOwnProperty(key))
			return false;
	}
	return true;
}

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

function exportMTGA(arr) {
	let str = "";
	for (c of arr) {
		let set = c.set.toUpperCase();
		if (set == "DOM") set = "DAR"; // DOM is called DAR in MTGA
		let name = c.printed_name[app.Language];
		let idx = name.indexOf('//');
		if (idx != -1)
			name = name.substr(0, idx - 1);
		str += `1 ${name} (${set}) ${c.collector_number}\n`
	}
	return str;
}