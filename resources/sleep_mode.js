/**
 * Created by Titan on 22/12/2016.
 */

// Detect sleep time and sleep screen

var info = {begin:6.75,end:0.5};

function toDecimal(date){
    return date == null ? 0 : (date.getHours() + date.getMinutes()/60);
}

function Range(begin,end){
    this.contain = function(decimalDate){
        return decimalDate >= begin && decimalDate < end;
    }
}

var Sleeper = {
    ranges:{
        list:[],
        dayMode:true,
        awakeDuration:0,
        sleepDuration:0,
        add:function(range){
            this.list.push(range);
        },
        isInside:function(decimaldate){
            for(var i = 0 ; i < this.list.length ; i++){
                if(this.list[i].contain(decimaldate)){
                    return true;
                }
            }
            return false;
        },
        //  Compute time to next sleep (if isUp = true) or time to still sleeping
        computeSleep:function(isUp,date){
            var sleep = 0;
            if(isUp){
                if(this.dayMode){
                    sleep = info.end - date;
                }else{
                    if(date < info.end){
                        sleep = info.end - date;
                    }else{
                        sleep = 24 - date + info.end;
                    }
                }
            }else{
                if(info.begin < info.end){
                    if(date < info.begin){
                        sleep = info.begin-date;
                    }else{
                        sleep = 24 - date + info.begin;
                    }
                }else{
                    sleep = info.begin - date;
                }
            }
            return {hours:sleep,seconds:sleep*3600,ms:sleep*3600*1000,toString:function(){return this.hours + " h / " + this.seconds + " min / " + this.ms + " ms";}};
        },
        init:function(){
            if(info.begin < info.end){
                this.add(new Range(info.begin,info.end));
                this.awakeDuration = ((info.end - info.begin) * 3600*1000);
            }else{
                this.dayMode = false;
                this.add(new Range(info.begin,24));
                this.add(new Range(0,info.end));
                this.awakeDuration = ((24 - info.begin + info.end) * 3600*1000);
            }
            this.sleepDuration = 24 * 3600*1000 - this.awakeDuration;
        }
    },
    init:function(){
        this.ranges.init();
        var date = toDecimal(new Date());
        var stayWakeUp = this.ranges.isInside(date);
        var sleep = this.ranges.computeSleep(stayWakeUp,date);
        if(stayWakeUp){
            setTimeout(function(){Sleeper.askForSleep(Sleeper.ranges.sleepDuration);},sleep.ms)
        }else{
            this.askForSleep(sleep.ms)
        }
        return this;
    },
    askForSleep:function(sleepTime){
        // Compute sleepTime if absent
        sleepTime = sleepTime || this.ranges.computeSleep(false,toDecimal(new Date())).ms;
        var begin = new Date().getTime();
        $('#idWarning').show();
        var context = {
            hasMove:false,
            move:function(){
                this.hasMove = true;
                $('#idWarning').hide();
                $('#idWarning').unbind('mousemove');
                // retry in 30 minutes, delay sleepTime
                sleepTime -= 1800*1000 + (new Date().getTime() - begin);
                if(sleepTime > 0){
                    console.log(new Date(),"Wait 30 minutes for sleeping",sleepTime,(sleepTime/(1000*60)),"min");
                    setTimeout(function(){Sleeper.askForSleep(sleepTime);},1800*1000);
                }
            },
            init:function(){
                var _self = this;
                setTimeout(function(){
                    if(!_self.hasMove){
                        $('#idWarning').unbind('mousemove').hide();
                        sleepTime = sleepTime - (new Date().getTime() - begin);
                        Sleeper.sleep(sleepTime);
                    }
                },60*1000);
                return this;
            }
        }.init();
        $('#idWarning').unbind('mousemove').bind('mousemove',function(e){
            if(this._detectMove()){
                context.move();
            }
            this.x = e.clientX;
            this.y = e.clientY;
        });
    },
    _detectMove:function(){
        return this.x != null && this.y != null && (this.x != e.clientX || this.y != e.clientY);
    },
    sleep:function(sleepTime){
        sleepTime = sleepTime || this.ranges.sleepDuration
        console.log(new Date(),"Sleep for",sleepTime/(1000*60),"min");
        // ask server to sleep
        TimeEventManager.pauseUntil(sleepTime);
        Clock.pause();
        // Detect mouse move on screen
        $('.gallery').unbind('mousemove').bind('mousemove',function(e){
            // Wake up sleep
            if(this._detectMove()){
                $('#idWarning').hide();
                $('.gallery').unbind('mousemove');
                // Wake up only 30 minutes and ask again if necessary. Compute if after 30 minutes, has to sleep
                var time = Sleeper.ranges.computeSleep(false,toDecimal(new Date()));
                var restSleep = time.ms - 1800*1000;
                if(restSleep > 0){
                    // ask again after 30 minutes. Compute the sleep time again
                    Sleeper.wakeUp(1800*1000);
                }else{
                    // No need, extend awakeDuration
                    Sleeper.wakeUp(Sleeper.ranges.awakeDuration - restSleep);
                }
            }
            this.x = e.clientX;
            this.y = e.clientY;
        });
        $.ajax({url:'/turnOff'});
        // Wake up after sleep duration
        setTimeout(function(){Sleeper.wakeUp();},sleepTime);
    },
    // use awakeDuration to override default value
    // Stay wakeup duration awakeDuration
    wakeUp:function(awakeDuration){
        $('.gallery').unbind('mousemove');
        awakeDuration = awakeDuration || this.ranges.awakeDuration;
        console.log(new Date(),"Wake up for",awakeDuration/(1000*60),"min");
        Clock.restart();
        // ask server to wakeup
        TimeEventManager.cancelPause();
        $.ajax({url:'/turnOn'});
        // Sleep after awake duration
        setTimeout(function(){Sleeper.askForSleep();},awakeDuration);
    }

}.init();