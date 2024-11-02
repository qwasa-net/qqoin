//
const { beginCell, Cell, contractAddress, toNano, storeStateInit, Address, Dictionary } = require("@ton/ton");
const { sha256_sync } = require("@ton/crypto");
const fs = require("fs");
const qs = require("querystring");

//
const QQOKEN_CONTRACT_FILE = "compiled/qqoken.boc64";
const QQOLLECTION_CONTRACT_FILE = "compiled/qqollection.boc64";
//
const DEPLOY_FEE = toNano(0.11);
const TRANSFER_FEE = toNano(0.06);
const MINT_FEE = toNano(0.06);

//
const WORKCHAIN = 0;
let TEST_NET = true;

let QQOKEN_BASE_URI = "https://qqoken.qwasa.net/~";

let QQOLLECTION_ON_DATA = {
    "name": "qqollection",
    "description": "·· qqoin qqoken qqollection ··",
    "image": "https://qqoken.qwasa.net/qqoken.webp"
};

let QQOKEN_ON_DATA = {
    "name": "qqoken",
    "description": "qqoin qqoken",
    "image": "https://qqoken.qwasa.net/qqoken.webp"
};

//
function build_offchain_meta(data) {
    const data_cell = beginCell()
        .storeUint(1, 8)
        .storeStringTail(String(data))
        .endCell();
    return data_cell;
}

//
function build_onchain_meta(data) {

    let metadata = Dictionary.empty(Dictionary.Keys.BigUint(256), Dictionary.Values.Cell());

    // name -- UTF8 string. Identifies the asset.
    if (data.name) {
        metadata.set(
            BigInt('0x' + sha256_sync("name").toString('hex')),
            beginCell().storeUint(0, 8).storeStringTail(data.name).endCell()
        );
    }

    // description -- UTF8 string. Describes the asset.
    if (data.description) {
        metadata.set(
            BigInt('0x' + sha256_sync("description").toString('hex')),
            beginCell().storeUint(0, 8).storeStringTail(data.description).endCell()
        );
    }

    // image -- ASCII string. A URI pointing to a resource with mime type image.
    if (data.image) {
        metadata.set(
            BigInt('0x' + sha256_sync("image").toString('hex')),
            beginCell().storeUint(0, 8).storeStringTail(data.image).endCell()
        );
    }

    // image_data -- Either binary representation of the image for onchain layout or base64 for offchain layout.
    if (data.image_data) {
        metadata.set(
            BigInt('0x' + sha256_sync("image_data").toString('hex')),
            beginCell().storeUint(0, 8).storeStringTail(data.image_data).endCell()
        );
    }

    // uri --  ASCII string. Used by "Semi-chain content layout". A URI pointing to JSON document with metadata.
    if (data.uri) {
        metadata.set(
            BigInt('0x' + sha256_sync("uri").toString('hex')),
            beginCell().storeUint(0, 8).storeStringTail(data.uri).endCell()
        );
    }

    const data_cell = beginCell()
        .storeUint(0, 8)
        .storeDict(metadata)
        .endCell();

    return data_cell;
}

//
function build_code_cell(boc_file) {
    const code_boc64 = fs.readFileSync(boc_file);
    let code_boc = Buffer.from(code_boc64.toString(), "base64");
    const code_cell = Cell.fromBoc(code_boc)[0];
    return code_cell;
}

//
function deploy_qqollection(qqollection_id, auth_addr, content) {

    console.log("@deploy_qqollection", qqollection_id, auth_addr);

    //
    const code_cell = build_code_cell(QQOLLECTION_CONTRACT_FILE);

    //
    const data_cell = beginCell()
        .storeAddress(auth_addr)
        .storeUint(qqollection_id, 64)
        .storeUint(0, 64)
        .storeRef(content)
        .storeRef(build_code_cell(QQOKEN_CONTRACT_FILE))
        .endCell();

    const state_init = { code: code_cell, data: data_cell };

    const stateInitBuilder = beginCell();
    storeStateInit(state_init)(stateInitBuilder);
    const state_init_cell = stateInitBuilder.endCell();

    const address = contractAddress(WORKCHAIN, state_init);
    console.log("qqollection addr:", address);

    print_transfer_address("deploy qqollection", address, DEPLOY_FEE, state_init_cell, null);

    return address;

}

//
function deploy_qqoken_from_qqollection(qqollection_addr, qqoken_id) {

    console.log("@deploy_qqoken_from_qqollection", qqollection_addr, qqoken_id);

    //
    let qqoken_addr = calculate_empty_qqoken_address(qqollection_addr, qqoken_id);
    console.log("qqoken addr:", qqoken_addr);

    //
    const data_cell = beginCell()
        .storeUint(0xffaaaaaa, 32) // mint op
        .storeUint(0, 64)
        .storeUint(qqoken_id, 64) // qqoken_id
        .storeCoins(MINT_FEE.toString(10))
        .endCell();

    print_transfer_address("mint qqoken", qqollection_addr, MINT_FEE, null, data_cell);

    return qqoken_addr;

}

//
function calculate_empty_qqoken_address(qqollection_addr, qqoken_id) {

    const code_cell = build_code_cell(QQOKEN_CONTRACT_FILE);
    const data_cell = beginCell()
        .storeUint(qqoken_id, 64)
        .storeAddress(qqollection_addr)
        .endCell();

    const state_init = { code: code_cell, data: data_cell };
    const address = contractAddress(WORKCHAIN, state_init);

    return address;

}

//
function transfer_qqoken(qqoken_addr, qqoken_id, owner_addr, content) {

    console.log("@transfer_qqoken", qqoken_addr, qqoken_id, owner_addr);

    const data_cell = beginCell()
        .storeUint(0x5fcc3d14, 32)
        .storeUint(qqoken_id, 64)
        .storeAddress(owner_addr)
        .storeAddress(owner_addr)
        .storeRef(content)
        .endCell();

    print_transfer_address("transfer qqoken", qqoken_addr, TRANSFER_FEE, null, data_cell);

}

//
function build_qqoken_content(qqoken_id, value) {
    let meta = JSON.parse(JSON.stringify(QQOKEN_ON_DATA));
    if (value) {
        meta["description"] = `·· qqoken #${qqoken_id} 🗲 value: ${value} qqoins ··`;
        meta["name"] = `qqoken #${qqoken_id}`;
    }
    meta["uri"] = QQOKEN_BASE_URI + qqoken_id;
    console.log("meta:", meta);
    return meta;
}

function take_excesses(qqoken, qqoken_id) {

    console.log("@take_excesses", qqoken);
    let qqoken_addr = Address.parse(qqoken);

    const data_cell = beginCell()
        .storeUint(0x5f3b5b3d, 32)
        .storeUint(qqoken_id, 64)
        .endCell();

    print_transfer_address("take excesses", qqoken_addr, TRANSFER_FEE, null, data_cell);

}

function print_transfer_address(text, address, amount, init, bin) {

    let params = {};
    if (amount) {
        params["amount"] = amount.toString(10);
    }
    if (bin) {
        params["bin"] = bin.toBoc({ idx: false }).toString("base64");
    }
    if (init) {
        params["init"] = init.toBoc({ idx: false }).toString("base64");
    }
    if (text) {
        params["text"] = text;
    }

    let link = "ton://transfer/" +
        address.toString({ testOnly: TEST_NET }) +
        "?" + qs.stringify(params);

    console.log(`=== ${text} ===`, link);

}

function parse_cmd_arguments() {

    const args = process.argv;
    let auth_addr = null;
    let owner_addr = null;
    let value = 0;
    let qqolection_id = 0;
    let qqoken_id = 0;
    let testnet = TEST_NET;

    args.forEach((arg, index) => {
        if (arg === '--auth' && args[index + 1]) {
            auth_addr = args[index + 1];
        }
        if (arg === '--owner' && args[index + 1]) {
            owner_addr = args[index + 1];
        }
        if (arg === '--value' && args[index + 1]) {
            value = args[index + 1];
        }
        if (arg === '--qqolection-id' && args[index + 1]) {
            qqolection_id = Number(args[index + 1]);
        }
        if (arg === '--qqoken-id' && args[index + 1]) {
            qqoken_id = Number(args[index + 1]);
        }
        if (arg === '--mainnet') {
            testnet = false;
        }

    });

    return {
        auth_addr: auth_addr,
        owner_addr: owner_addr,
        value: value,
        qqolection_id: qqolection_id,
        qqoken_id: qqoken_id,
        testnet: testnet
    };
}

function do_the_do(params) {

    // [1] deploy collection
    console.log("\n[1] deploy qqollection");
    let auth_addr = Address.parse(params.auth_addr);
    let qqollection_id = params.qqolection_id;
    let qqollection_content = build_onchain_meta(QQOLLECTION_ON_DATA);
    let qqollection = deploy_qqollection(
        qqollection_id,
        auth_addr,
        qqollection_content
    );

    // [2] mint empty qqoken
    console.log("\n[2] mint empty qqoken");
    let qqoken_id = params.qqoken_id;
    qqoken_addr = deploy_qqoken_from_qqollection(qqollection, qqoken_id);

    // [3] transfer qqoken
    if (params.value && params.owner_addr) {
        console.log("\n[3] transfer qqoken");
        let owner_addr = Address.parse(params.owner_addr);
        let value = params.value;
        let qqoken_content = build_onchain_meta(build_qqoken_content(qqoken_id, value));
        transfer_qqoken(qqoken_addr, qqoken_id, owner_addr, qqoken_content);
    }

}

let args = parse_cmd_arguments();
TEST_NET = args.testnet;
console.log(args);
do_the_do(args);
