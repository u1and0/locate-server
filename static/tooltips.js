function toggleMenu(elem){
  const obj = document.getElementById(elem).style;
  obj.display=(obj.display=='none')?'block':'none';
}

// Loading spinner
$(function(){
  var loader = $('.loader-wrap');

  //ページの読み込みが完了したらアニメーションを非表示
  $('#result').on('load',function(){
    loader.fadeOut();
  });

  //ページの読み込みが完了してなくても3秒後にアニメーションを非表示にする
  setTimeout(function(){
    loader.fadeOut();
  },5000);
});
