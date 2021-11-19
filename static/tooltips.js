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
  },1000);
});

// $("#result").pagination({
//   itemElement: "> td"
// });

// $('#pagination-container').pagination({
//     dataSource: [1, 2, 3, 4, 5, 6, 7],
//     callback: function(data, pagination) {
//         var html = simpleTemplating(data);
//         $('#result').html(html);
//     }
// });
//
// function simpleTemplating(data) {
//     var html = '<ul>';
//     $.each(data, function(index, item){
//         html += '<li>'+ item +'</li>';
//     });
//     html += '</ul>';
//     return html;
// }
