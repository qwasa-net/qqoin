const qqoinapp = {

    tg: window.Telegram.WebApp,
    clicker_el: null,
    score_el: null,

    score_total: 0,
    clicker_energy: 0,
    clicker_energizer: 1,
    state: 0, // 0=init, 1=running, 2=crashed
    tix: null,
    tix_countdown: 0,
    user_id: 0,
    hello: null,
    xyz: null,

    TIX: 30,
    API_BASE_URL: "https://qqoin-api.qqoin.qq/api/",
    CONSOLE_LOGGING: window.location.protocol === "file:",

    log: function (...args) {
        if (this.CONSOLE_LOGGING) {
            console.log("qqoin:", ...args);
        }
    },

    init: function () {
        this.log("init");
        this.state = 0;
        this.score_el = document.getElementById('score');
        this.score_update_el = document.getElementById('score_update');
        this.clicker_el = document.getElementById('clicker');
        this.clicker_el.addEventListener("mousedown", this.handle_clicker_clicked);
        try { this.tg.ready(); this.tg.expand(); } catch (e) { }
        try { this.user_id = this.tg.initDataUnsafe.user.id; } catch (e) { this.user_id = 0; }
        this.get_data_init((data) => { this.init_with_data(data); }, (err) => { this.init_with_data(); });
    },

    init_with_data: function (data) {
        this.log("init_with_data", data);
        if (data) {
            this.score_total = data.score || 0;
            this.hello = data.hello || Math.random().toString(36).substring(2, 15);
        } else {
            this.score_total = -1;
            this.state = -1;
        }
        this.update_clicker_label();
        this.update_score_label();
    },

    start: function () {
        this.state = 1;
        this.reset_labels();
        this.reset_clicker();
        this.start_tixer();
        this.log("started", tg.initData, tg.initDataUnsafe);
    },

    recover: function () {
        this.clicker_el.classList.remove("soon");
        this.clicker_el.classList.remove("danger");
        this.state = 3;
        this.update_clicker_label();
        setTimeout((ev) => { this.start(); }, 500);
    },

    energize: function () {
        this.stop_tixer();
        let energy = Math.floor(Number(this.clicker_energy) || 0);
        this.score_total += energy;
        this.post_data_updates(energy);
        this.clicker_el.classList.remove("soon");
        this.clicker_el.classList.add("success");
        this.update_clicker_label();
        this.update_score_label(`+${energy}`);
        setTimeout((ev) => { this.start(); }, 500);
    },

    deenergize: function () {
        this.stop_tixer();
        let energy = Math.floor(Number(this.clicker_energy) || 0);
        this.score_total -= energy;
        this.post_data_updates(-energy);
        this.state = 2;
        this.clicker_el.classList.remove("soon");
        this.clicker_el.classList.add("danger");
        this.update_clicker_label();
        this.update_score_label(`-${energy}`);
    },

    update_score_label: function (update) {
        this.score_el.innerText = String(this.score_total);
        if (update) {
            this.score_update_el.innerText = String(update);
            this.score_update_el.classList.add("slide-away");
        } else {
            this.score_update_el.innerText = "";
        }
    },

    update_clicker_label: function () {
        let energy = Math.floor(Number(this.clicker_energy) || 0);
        if (this.state === 0) {
            this.clicker_el.innerText = "qo!";
        } else if (this.state === -1) {
            this.clicker_el.innerText = "error!!";
        } else if (this.state === 2) {
            this.clicker_el.innerText = String(-energy);
        } else if (this.state === 3) {
            this.clicker_el.innerText = "··";
        } else if (this.clicker_energy <= 0) {
            this.clicker_el.innerText = "";
        } else {
            this.clicker_el.innerText = String(energy);
            if (this.tix_countdown < this.TIX / 1.5) {
                this.clicker_el.classList.add("soon");
            }
        }
    },

    fill_xyz: function (e) {
        if (!e) { this.xyz = null; return; }
        function asciisum(s) {
            return String(s).split("").reduce((a, c) => a + c.charCodeAt(0), 0);
        }
        let now = Date.now();
        xyz = [
            now / 1000, now % 1000,
            e.clientX, e.clientY, e.offsetX, e.offsetY, e.pageX, e.pageY,
            asciisum(e.type),
            asciisum(e.target && e.target.getAttribute("id"))
        ];
        this.xyz = xyz.map((v) => Math.floor(Number(v) || 0));
    },

    handle_clicker_clicked: function (e) {
        if (qqoinapp.state == 0) {
            qqoinapp.start();
            return;
        }
        qqoinapp.fill_xyz(e);
        if (qqoinapp.state == 2) {
            qqoinapp.recover();
            return;
        }
        if (qqoinapp.clicker_energy >= 1) {
            qqoinapp.energize();
        }
    },

    reset_labels: function () {
        this.clicker_el.classList.remove("soon");
        this.clicker_el.classList.remove("danger");
        this.clicker_el.classList.remove("success");
        this.score_update_el.innerText = "";
        this.score_update_el.classList.remove("slide-away");
    },

    reset_clicker: function () {
        let rnd = Math.random();
        if (rnd > 0.75) { rnd *= 3; }
        this.tix_countdown = Math.floor(this.TIX + 3 * this.TIX * rnd);
        this.clicker_energy = 0;
        this.clicker_energizer = 0;
    },

    start_tixer: function () {
        clearInterval(this.tixer);
        this.tixer = setInterval(this.do_tix, 1000 / this.TIX);
    },

    stop_tixer: function () {
        clearInterval(this.tixer);
    },

    add_energy: function () {
        this.clicker_energizer++;
        this.clicker_energy += Math.max(Math.log(this.clicker_energizer) - 2, 4 / this.TIX);
    },

    do_tix: function () {
        if (qqoinapp.state != 1) return;
        qqoinapp.tix_countdown--;
        if (qqoinapp.tix_countdown <= 0) {
            qqoinapp.deenergize();
            return;
        } else {
            qqoinapp.add_energy();
            qqoinapp.update_clicker_label();
        }
    },

    post_data_updates: function (energy, cb_success, cb_error) {
        let data = {
            "init": this.tg.initData,
            "uid": this.user_id,
            "hello": this.hello,
            "e": energy || 0,
            "s": this.score_total,
            "xyz": this.xyz,
        };
        const url = this.API_BASE_URL + "taps/";
        this.log("post_data_updates", url, data);
        try {
            fetch(url, {
                method: 'POST',
                body: JSON.stringify(data),
                headers: { 'Content-Type': 'application/json' }
            }).then((rsp) => {
                rsp.json().then(
                    (data) => {
                        this.log(data);
                        if (cb_success) { cb_success(data); }
                    }
                ).catch((error) => {
                    this.log(error);
                    if (cb_error) { cb_error(error); }
                });
            }).catch((error) => {
                this.log(error);
                if (cb_error) { cb_error(error); }
            }).finally(() => {
                this.fill_xyz(null);
            });
        } catch (error) {
            this.log(error);
        }
    },

    get_data_init: function (cb_success, cb_error) {
        let data = {
            "init": this.tg.initData,
            "uid": this.user_id,
        };
        const url = this.API_BASE_URL + "taps/";
        try {
            fetch(url, {
                method: 'POST', // wanna GET, not cry
                body: JSON.stringify(data),
                headers: { 'Content-Type': 'application/json' }
            }).then((rsp) => {
                rsp.json().then(
                    (data) => {
                        this.log(data);
                        if (cb_success) { cb_success(data); }
                    }
                ).catch((error) => {
                    this.log(error);
                    if (cb_error) { cb_error(error); }
                });
            }).catch((error) => {
                this.log(error);
                if (cb_error) { cb_error(error); }
            });
        } catch (error) {
            this.log(error);
        }
    },

};

const tg = window.Telegram.WebApp;
if (tg) {
    if (window.QQOIN_API_BASE_URL) {
        qqoinapp.API_BASE_URL = window.QQOIN_API_BASE_URL;
    }
    window.addEventListener("load", (win, ev) => { qqoinapp.init(); });
}


