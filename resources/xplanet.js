/**
 * Created by Titan on 24/10/2016.
 */


// Display image about earth, rolling earth and moon phase
var XPlanet = {
    earth:null,
    moon:null,
    autoInit:function(){
        if($('.planet').length > 0){
            this.earth = $('img.earth','.planet');
            this.moon = $('img.moon','.planet');
            this.init();
        }
        return this;
    },
    init:function(){
        var _self = this;
        TimeEventManager.add(PatternManager.every10Minutes(),function(){_self.setEarth();},true);
        TimeEventManager.add(PatternManager.everyHours(),function(){_self.setMoon();},true);
    },
    // Load image of earth and display. Use local time (round floor to 10 minutes)
    setEarth:function(){
        var d = new Date();
        // Animate case at hour begin
        if(d.getMinutes() == 0){
            console.log("Show hour")
            var imgSrc = "/image?format=gif&planet=earth&date=" + Clock._pad(d.getHours()) + "&rand=" + Math.random();
            this.earth.attr("src",imgSrc);
            // Change animate to fixe after one minute
            setTimeout(function(){XPlanet.setEarth();},60000 - (d.getSeconds()*1000));
        }else{
            var time = Clock._pad(d.getHours()) + Clock._pad(Math.floor(d.getMinutes()/10) * 10);
            // case earth at xx00 and anime gif
            var imgSrc = "/image?format=jpg&planet=earth&date=" + time;
            this.earth.attr("src",imgSrc);
        }
    },
    // load the moon image
    setMoon:function(){
        this.moon.attr("src","/image?format=jpg&planet=moon&rand=" + Math.random());
        var d = new Date();
        // compute rise and set
        var moon = SunCalc.getMoonTimes(d,48.8,2.3);
        if(d > moon.rise && d > moon.set ){
            // Reload for next date
            d.setDate(d.getDate()+1);
            moon = SunCalc.getMoonTimes(d,48.8,2.3);
            moon.rise < moon.set ? this._switch('moonrise','moonset') : this._switch('moonset','moonrise');
        }
        this._setMoon(moon);
        if(d < moon.rise && d < moon.set){
            // write the closest first
            if(moon.rise < moon.set){
                this._switch('moonrise','moonset');
            }else {
                this._switch('moonset','moonrise');
            }
            moon.rise < moon.set ? this._switch('moonrise','moonset') : this._switch('moonset','moonrise');
        }
        if(d > moon.rise && d < moon.set){
            this._switch('moonrise','moonset');
        }
        if(d < moon.rise && d > moon.set){
            this._switch('moonset','moonrise');
        }
    },
    _setMoon:function(moon){
        $('div.planet span.moonrise > span.clock').html(this._formatTime(moon.rise));
        $('div.planet span.moonset > span.clock').html(this._formatTime(moon.set));
    },
    _switch:function(left,right){
        $('div.planet span.' + left).removeClass("moonright").addClass("moonleft");
        $('div.planet span.' + right).removeClass("moonleft").addClass("moonright");
    },
    _formatTime:function(date){
        if(date == null){
            return "";
        }
        return (date.getHours()< 10?"0":"") + date.getHours() + "h" + (date.getMinutes() < 10 ? "0":"") + date.getMinutes();
    }
}.autoInit();