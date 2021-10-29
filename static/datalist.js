$("#query").on('input', function () {
    var val = this.value;
    if($('#searched-words option').filter(function(){
        return this.value.toUpperCase() === val.toUpperCase();
    }).length) {
        //send ajax request
        alert(this.value);
    }
});
