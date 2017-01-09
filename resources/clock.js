// Manage the clock and the differents event
// Needed : change some image every 10 minutes (earth), some times every one hour (animated earth), every day (moon).

//Create time pattern for each usage. Usefull functions on date / time are presents
// A pattern is based on time (HHMMS)
var PatternManager = {
    locale:'fr-fr',
    thirtySeconds:new RegExp("^[0-9]{4}[03]$"),
    minute:new RegExp("^[0-9]{4}0$"),
    tenMinutes:new RegExp("^[0-9]{3}00$"),
    twentyMinutes:new RegExp("^[0-9]{2}((20)|(40)|(00))0$"),
    thirtyMinutes:new RegExp("^[0-9]{2}((30)|(00))0$"),
    hour:new RegExp("^[0-9]{2}000$"),
    day:new RegExp("^00000$"),
    every30Seconds:function(){
        return this.thirtySeconds;
    },
    everyMinute:function(){
        return this.minute;
    },
    every10Minutes:function(){
        return this.tenMinutes;
    },
    every20Minutes:function(){
        return this.twentyMinutes;
    },
    every30Minutes:function(){
        return this.thirtyMinutes;
    },
    everyHours:function(){
        return this.hour;
    },
    everyDay:function(){
        return this.day;
    },
    // Check if specified time (HHMM) check the pattern
    check:function(time,pattern){
        return pattern.test(time);
    },
    // return a format date as HHMM
    getFormatHour:function(){
        var d = new Date().toLocaleTimeString();
        return d.substring(0,2) + d.substring(3,5) + d.substring(6,7);
    },
    // return a date like : samedi 10 novembre 2016
    getFormatDate:function(){
        var d = new Date();
        return this._firstLetterCaps(d.toLocaleString(this.locale, { weekday: "long"})) + " "
            + ((d.getDate() < 10) ? "0" :"") + d.getDate() + " "
            + this._firstLetterCaps(d.toLocaleString(this.locale, { month:"long" })) + " "
            + d.getFullYear();
    },
    _firstLetterCaps:function(name){
        if(name == null || name == ""){return "";}
        return name.substring(0,1).toUpperCase() + name.substring(1);
    }
};


// initialDelay is used to wait different duration at first launch. Used only the first time; If not present, delay used instead
window.setCorrectingInterval = ( function( func, delay, initialDelay ) {
    var instance = {func:func,delay:delay,started:false,target:initialDelay || delay,startTime:new Date().valueOf(),running:true};
    instance.stop = function(){
        this.running = false;
    }
    instance.start=function(waitInMs){
        this.running = true;
        this.started = false;
        this.startTime = new Date().valueOf();
        this.target = waitInMs;
        setTimeout(tick,waitInMs);
    }
    function tick( delay ) {
        if(!instance.running){
            return;
        }
        if ( ! instance.started ) {
            instance.started = true;
            setTimeout( tick, initialDelay || delay );
            initialDelay = null;
            // First launch, return instance
            return instance;
        } else {
            var elapsed = new Date().valueOf() - instance.startTime;
            var adjust = instance.target - elapsed;
            instance.func();
            instance.target += instance.delay;
            setTimeout( tick, instance.delay + adjust );
        }
    };
    return tick( delay );
} );


// Launch event at specific time
var TimeEventManager = {
    events:[],  // store Event
    instanceRunning:null,
    // Make test every minute
    init:function(){
        // Launch one time at begin and search the next full minute (hhmm:00) : 60 - actual seconds)
        var sec = new Date().getSeconds();
        this.instanceRunning = setCorrectingInterval(function(){TimeEventManager.getEventsToLaunch();},1000*30,1000*(sec <= 30 ? 30 - sec : 60 - sec));
        return this;
    },
    getEventsToLaunch:function(){
        var time = PatternManager.getFormatHour();
        this.events.forEach(function(e){
            if(PatternManager.check(time,e.pattern)){
                setTimeout(e.actionFct);
            }
        });
    },
    // if launchImmediate is true, launch function when adding
    add:function(pattern,actionFct,launchImmediate){
        this.events.push(new Event(pattern,actionFct));
        if(launchImmediate){
            actionFct();
        }
    },
    // Put manager on standby until next specific time
    pauseUntil:function(waitTimeInMS){
        var _self = this;
        this.instanceRunning.stop();
        setTimeout(function(){
            _self.instanceRunning.start(0);
        },waitTimeInMS);
    },
    // Restart manager
    cancelPause:function(){
        var sec = new Date().getSeconds();
        var wait = 1000*(sec <= 30 ? 30 - sec : 60 - sec);
        this.instanceRunning.start(wait);
    }
}.init();

// pattern is generate by PatternManager
function Event(pattern,actionFct) {
    this.pattern = pattern;
    this.actionFct = actionFct;
}

// Change date every second
// Mecanism of autoinit, if a specific class exist
var Clock = {
    timeDiv:null,
    dateDiv:null,
    secondInterval:null,
    // Try to init clock by searching specific class in html
    autoInit:function(){
        if($('div.clock').length != 0 ){
            this.timeDiv = $('.time','div.clock');
            this.dateDiv = $('.date','div.clock');
            this.init();
        }
        return this;
    },
    init:function(){
        this.setFull();
        this.start();
    },
    setFull:function(){
        this.setDate();
        this.setTime();
        this.setSeconds();
    },
    setDate:function(){
        this.dateDiv.html(PatternManager.getFormatDate());
    },
    setTime:function(){
        var d = new Date();
        $('.hm',this.timeDiv).html( this._pad(d.getHours()) + ":" + this._pad(d.getMinutes()));
    },
    setSeconds:function(){
        $('.seconds',this.timeDiv).html(this._pad(new Date().getSeconds()));
    },
    // prefix number with zero if < 10
    _pad:function(value){
        return ((value < 10)?"0":"") + value;
    },
    // Start the management
    start:function(){
        TimeEventManager.add(PatternManager.everyMinute(),function(){Clock.setTime();});
        TimeEventManager.add(PatternManager.everyDay(),function(){Clock.setDate();});
        // update Seconds
        this.secondInterval = setInterval(function(){Clock.setSeconds();},1000);
    },
    pause:function(){
        if(this.secondInterval!=null){
            clearInterval(this.secondInterval);
            this.secondInterval = null;
        }
    },
    restart:function(){
        this.secondInterval = setInterval(function(){Clock.setSeconds();},1000);
    }
}.autoInit();