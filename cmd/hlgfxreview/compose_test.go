package main

// Exercise saved collections and actual DOM placement with the same catalog
// and image snapshots used by the single-asset browser checks.
const composerSmoke = `
(()=>{
 const assert=(x,m)=>{if(!x)throw Error(m)};
 const fail=e=>{document.body.dataset.smoke='FAIL: '+e.message;document.getElementById('message').textContent=e.stack;};
 document.body.dataset.smoke='running';
 const change=(id,value)=>{const e=document.getElementById(id);e.value=value;e.dispatchEvent(new Event(id==='search'?'input':'change',{bubbles:true}));};
 const clickCell=(row,col)=>document.querySelector('.hex[data-row="'+row+'"][data-col="'+col+'"]').click();
 try {
  assert(scene().cells.filter(c=>c.layers.some(l=>l.asset.startsWith('background/hq'))).length===7,'complete HQ');
  const n=cell(scene(),1,1).layers[0],nw=cell(scene(),2,1).layers[0],center=cell(scene(),3,1).layers[0];
  assert(n.asset==='background/hqN'&&nw.asset==='background/hqNW'&&center.asset==='background/hqC','HQ part bindings');
  const p=origin(1,1),q=origin(2,1);assert(p.x-q.x===71&&q.y-p.y===40,'compound registration');
  for(const [template,count]of[['hq-summer',7],['hq-winter',7],['depot-summer',4],['depot-winter',4],['factory-winter',4]]){
   const s=makeScene(template);validate({schema_version:1,name:'templates',scenes:[s]});
   assert(s.cells.filter(c=>c.layers.some(l=>!l.asset.startsWith('background/grass'))).length===count,'template '+template);
  }
  const roadScene=makeScene('road-unit');assert([0,2,4,6].every(row=>cell(roadScene,row,1).layers.some(l=>l.asset==='infrastructure/road36')),'connected vertical road template');
  change('template','grass');document.getElementById('add-scene').click();
  assert(collection.scenes.length===2,'user scene added');
  change('scene-name','My combined scene');change('collection-name','My review collection');
  assert(scene().name==='My combined scene','scene name');
  document.querySelector('[data-asset="infrastructure/road14"]').click();clickCell(2,1);
  document.querySelector('[data-asset="sprite/infantry"]').click();change('variant','ec/se-spent');clickCell(2,1);
  let c=cell(scene(),2,1);assert(c.layers.length===3,'terrain road and unit stack');
  let unit=c.layers.find(l=>l.asset==='sprite/infantry');assert(unit.variant==='ec/se-spent','authored variant retained');
  const ims=Array.from(document.querySelectorAll('#views img')).filter(i=>i.dataset.row==='2'&&i.dataset.col==='1');
  assert(ims.map(i=>i.dataset.asset).join(',')==='background/grass,infrastructure/road14,sprite/infantry','global layer draw order');
  const im=ims.at(-1);assert(im.style.width==='96px'&&im.style.top==='87px','96px unit and -9px registration');
  assert(im.src.endsWith(assets.get('sprite/infantry').variants.find(v=>v.id==='ec/se-spent').images.game.url),'Lanczos game-size sprite');
  change('offset-y','3');assert(cell(scene(),2,1).layers.at(-1).dy===3,'placement offset');
  document.getElementById('visibility').click();assert(!cell(scene(),2,1).layers.at(-1).visible,'layer hidden');
  document.getElementById('undo').click();assert(cell(scene(),2,1).layers.at(-1).visible,'undo hide');
  document.getElementById('redo').click();assert(!cell(scene(),2,1).layers.at(-1).visible,'redo hide');document.getElementById('undo').click();
  const transfer=new DataTransfer();transfer.setData('text/plain','infrastructure/rail14');
  document.querySelector('.hex[data-row="2"][data-col="1"]').dispatchEvent(new DragEvent('drop',{dataTransfer:transfer,bubbles:true,cancelable:true}));
  assert(cell(scene(),2,1).layers.some(l=>l.asset==='infrastructure/rail14'),'drag/drop placement');
  change('view','wipe');document.getElementById('wipe').value='35';document.getElementById('wipe').dispatchEvent(new Event('input'));
  assert(document.querySelector('.layer-top').style.clipPath==='inset(0px 0px 0px 35%)','assembled wipe');
  const saved=JSON.stringify(collection),restored=validate(JSON.parse(saved));assert(JSON.stringify(restored)===saved,'collection round trip');
  const stored=JSON.parse(localStorage.getItem(storageKey));assert(stored.name==='My review collection'&&stored.scenes.length===2,'local saving');
  const bad=clone(collection);bad.scenes[0].cells[0].layers[0].asset='https://external.example/image.png';
  let rejected=false;try{importText(JSON.stringify(bad));}catch{rejected=true;}assert(rejected&&JSON.stringify(collection)===saved,'invalid import keeps collection');
  for(const mutate of [s=>s.rows=200,s=>s.cells.push(clone(s.cells[0])),s=>s.cells[0].layers[0].variant='unknown',s=>s.cells[0].layers[0].dx=1.5]){
   const bad=clone(collection);mutate(bad.scenes[0]);let rejected=false;try{validate(bad);}catch{rejected=true;}assert(rejected,'invalid scene rejected');
  }
  const large={schema_version:1,name:'Too large',scenes:[]};
  for(let i=0;i<50;i++){const s=makeScene('grass');s.id='big-'+i;s.columns=12;s.rows=21;s.cells=[];for(let row=0;row<21;row++)for(let col=0;col<12-row%2;col++)s.cells.push({row,col,layers:[layer('background/grass')]});large.scenes.push(s);}
  let largeRejected=false;try{validate(large);}catch{largeRejected=true;}assert(largeRejected,'export size remains importable');
  document.getElementById('duplicate').click();assert(collection.scenes.length===3,'duplicate scene');document.getElementById('delete-scene').click();assert(collection.scenes.length===2,'delete scene');
  importText(saved);assert(JSON.stringify(collection)===saved,'import restores scene collection');
  activeID=collection.scenes[1].id;refresh();change('tool','select');clickCell(2,1);
  change('columns','2');change('rows','2');document.getElementById('resize').click();assert(scene().columns===2&&scene().rows===2&&scene().cells.every(c=>inside(scene(),c.row,c.col)),'resize bounds');
  document.getElementById('undo').click();assert(scene().columns===4&&cell(scene(),2,1).layers.length===4,'undo resize');
  change('grid','no');assert(document.querySelectorAll('.outlines polygon').length===0&&!document.querySelector('.hex.selected'),'clean artwork view');
  assert(getComputedStyle(document.querySelector('.hex[aria-pressed="true"]')).backgroundColor==='rgba(0, 0, 0, 0)','selected cell does not obscure artwork');
  change('view','side');change('density','2');change('zoom','2');
  assert(document.querySelectorAll('#views figure').length===2,'assembled side by side');
  const currentUnit=Array.from(document.querySelectorAll('#views figure')).at(-1).querySelector('img[data-asset="sprite/infantry"]');
  assert(currentUnit.src.endsWith(assets.get('sprite/infantry').variants.find(v=>v.id==='ec/se-spent').images['2'].url),'2x unit artwork selected');
  change('view','current');change('zoom','1');
  const finish=()=>{try{
   assert(Array.from(document.querySelectorAll('#views img')).every(i=>i.complete&&i.naturalWidth>0),'composed image loading');
   const im=document.querySelector('#views img[data-asset="sprite/infantry"]');assert(im.naturalWidth===192,'2x unit artwork loaded');
   assert(document.documentElement.scrollWidth<=innerWidth+1,'responsive composer overflow');
   document.body.dataset.smoke='pass';
  }catch(e){fail(e);}};
  if(document.readyState==='complete')finish();else window.addEventListener('load',finish,{once:true});
 }catch(e){fail(e);}
})();`
