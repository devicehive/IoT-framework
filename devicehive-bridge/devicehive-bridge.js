/**
 * Created by demon on 4/26/15.
 */

var dbus = require('dbus-native');
var _ = require('lodash');

var systemBus = dbus.systemBus();


var DH_CLOUD_NS = 'com.devicehive.cloud';
var DH_CLOUD_PATH = '/com/devicehive/cloud';

var DH_BLE_NS = 'com.devicehive.bluetooth';
var DH_BLE_PATH = '/com/devicehive/bluetooth';


var proc_scan_start = function (ctx, id, params, cb){
    console.log('Starting Scan');
    ctx.discovered_peripherals = {};
    ctx.ble.ScanStart();
    cb(null);
};

var proc_scan_stop = function (ctx, id, params, cb){
    console.log('Stopping Scan');
    ctx.ble.ScanStop();
    var results = ctx.discovered_peripherals;
    cb(null, results);
};

var proc_scan = function (ctx, id, params, cb){
    var delay = params.timeout || 10; // 10 sec by default
    console.log('Starting Scan for ' + delay + ' sec');
    ctx.discovered_peripherals = {};
    ctx.ble.ScanStart();
    setTimeout(function () {
        proc_scan_stop(ctx, id, params, cb);
    }, delay*1000);
};

var on_connect_callbacks = {};
var CONNECT_WAIT_TIMEOUT = 10000; // 5sec

var ensure_connected = function (ctx, mac, cb) {
    ctx.ble.Connect(mac, function (err, ready) {

        if (err){
            return cb(err);
        }

        if (ready)
            return cb();

        // if not ready schedule execution for on connect
        on_connect_callbacks[mac] = on_connect_callbacks[mac] || [];
        on_connect_callbacks[mac].push(cb);

        console.log('Peripheral is not ready. Scheduling execution', mac);

        // set timeout
        setTimeout(function () {
            var idx = on_connect_callbacks[mac].indexOf(cb);
            if (idx > -1){
                console.log('Peripheral Connect Timeout', mac);
                on_connect_callbacks[mac].splice(idx, 1);
                cb('Peripheral Connect Timeout');
            }
        }, CONNECT_WAIT_TIMEOUT);

    });
};

var proc_write = function (ctx, id, params, cb){

    var mac = params.device;
    var characteristic = params.characteristic;
    var value = params.value;

    if (!mac) throw "Parameter missing: 'device'";
    if (!characteristic) throw "Parameter missing: 'characteristic'";
    if (!value) throw "Parameter missing: 'value'";

    console.log('Writing', mac, characteristic, value);

    ensure_connected(ctx, mac, function (err, callback) {
        if (err)
            return cb(err, null, callback);

        ctx.ble.GattWrite(mac, characteristic, value, function (err) {
            cb(err, null, callback);
        });
    });

};

var proc_read = function (ctx, id, params, cb){

    var mac = params.device;
    var characteristic = params.characteristic;

    console.log('Reading', mac, characteristic);

    ensure_connected(ctx, mac, function (err, callback) {
        if (err)
            return cb(err, null, callback);

        ctx.ble.GattRead(mac, characteristic, value, function (err, value) {
            cb(err, {value: value}, callback);
        });
    });

};


var proc_notifications_start = function (ctx, id, params, cb){

    var mac = params.device;
    var characteristic = params.characteristic;

    console.log('Starting notifications', mac, characteristic);

    ensure_connected(ctx, mac, function (err, callback) {
        if (err)
            return cb(err, null, callback);

        ctx.ble.GattNotifications(mac, characteristic, true, function (err) {
            cb(err, null, callback);
        });
    });

};


var proc_notifications_stop = function (ctx, id, params, cb){

    var mac = params.device;
    var characteristic = params.characteristic;

    console.log('Stopping notifications', mac, characteristic);

    ensure_connected(ctx, mac, function (err, callback) {
        if (err)
            return cb(err, null, callback);

        ctx.ble.GattNotifications(mac, characteristic, false, function (err) {
            cb(err, null, callback);
        });
    });

};

var proc_indications_start = function (ctx, id, params, cb){

    var mac = params.device;
    var characteristic = params.characteristic;

    console.log('Starting indications', mac, characteristic);

    ensure_connected(ctx, mac, function (err, callback) {
        if (err)
            return cb(err, null, callback);

        ctx.ble.GattIndications(mac, characteristic, true, function (err) {
            cb(err, null, callback);
        });
    });

};


var proc_indications_stop = function (ctx, id, params, cb){

    var mac = params.device;
    var characteristic = params.characteristic;

    console.log('Stopping indications', mac, characteristic);

    ensure_connected(ctx, mac, function (err, callback) {
        if (err)
            return cb(err, null, callback);

        ctx.ble.GattIndications(mac, characteristic, false, function (err) {
            cb(err, null, callback);
        });
    });

};


var handlers = {
    //"gatt/primary"      : proc_primary,
    "gatt/read"         : proc_read,
    "gatt/write"        : proc_write,
    //"gatt/connect"      : proc_connect,
    //"gatt/disconnect"   : proc_disconnect,
    //"gatt/characteristics"    : proc_characteristics,
    "gatt/notifications"      : proc_notifications_start,
    "gatt/notifications/stop" : proc_notifications_stop,
    "gatt/indications"        : proc_indications_start,
    "gatt/indications/stop"   : proc_indications_stop,
    "scan/start"        : proc_scan_start,
    "scan/stop"         : proc_scan_stop,
    "scan"              : proc_scan
};

var device_discovered = function (ctx, mac, name, rssi) {
    var peripheral = {name : name, address: mac, rssi: rssi};
    ctx.discovered_peripherals[mac] = peripheral;
    console.log('Peripheral Discovered', mac, name, rssi);
    ctx.cloud.SendNotification('PeripheralDiscovered', JSON.stringify(peripheral))
};

var process_queue = function(mac){
    var queue = on_connect_callbacks[mac];
    var f = queue.shift();
    f(null, _.curry(process_queue)(mac))
};

var device_connected = function (ctx, mac) {
    ctx.cloud.SendNotification('PeripheralConnected', JSON.stringify({address: mac}))

    var queue = on_connect_callbacks[mac];
    if (!queue) return;
    console.log('Processing pending operations', mac,  queue.length);
    _.curry(process_queue)(mac)();
};


var notification_received = function (ctx, type, mac, uuid, message) {
    console.log(type, mac, uuid, message);
    ctx.cloud.SendNotification(type, JSON.stringify({
        address: mac,
        characteristic: uuid,
        value: message,
        type: type
    }))
};

systemBus.getService(DH_CLOUD_NS).getInterface(DH_CLOUD_PATH, DH_CLOUD_NS, function(err, cloud) {

    if (err){
        console.log('Error '+DH_CLOUD_NS+' dbus service: ' + err);
        process.exit(1);
    }
    console.log('Connected to', DH_CLOUD_PATH);

    systemBus.getService(DH_BLE_NS).getInterface(DH_BLE_PATH, DH_BLE_NS, function(err, ble) {

        if (err){
            console.log('Error '+DH_BLE_NS+' dbus service: ' + err);
            process.exit(1);
        }
        console.log('Connected to', DH_BLE_PATH);

        var context = {
            ble: ble,
            cloud: cloud,
            discovered_peripherals: {}
        };


        // handle commands  from cloud and marshal to local handlers
        cloud.on ('CommandReceived', function (id, name, paramsstr) {
            console.log (id, name, paramsstr);

            setTimeout(function () {
                try {
                    var cmd = handlers[name];
                    var params = JSON.parse(paramsstr || '{}');

                    if (!cmd)
                        throw 'Unknown command: ' + name;

                    cmd(context, id, params, function (err, result, cb) {
                        if (err){
                            console.log('Update Command', id, name, 'Error', err);
                            cloud.UpdateCommand(id, 'Error', JSON.stringify({error: err, details: (result||{})}));
                            if (cb) cb();
                        } else {
                            console.log('Update Command', id, name, 'Success', result);
                            cloud.UpdateCommand(id, 'Success', JSON.stringify(result||{}));
                            if (cb) cb();
                        }
                    });

                } catch(e) {
                    console.log('Update Command', id, name, 'Error', e);
                    cloud.UpdateCommand(id, 'Error', JSON.stringify({error: e}));
                }
            }, 10);

        });


        ble.on('DeviceDiscovered', _.curry(device_discovered)(context));
        ble.on('DeviceConnected', _.curry(device_connected)(context));
        ble.on('NotificationReceived', _.curry(notification_received)(context, 'Notification'));
        ble.on('IndicationReceived', _.curry(notification_received)(context, 'Indication'));

        console.log('Listening ....');

    });
});


