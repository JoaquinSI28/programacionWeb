// Prueba de integración: requiere una instancia de prueba con las cuatro listas vacías.
// Crea y elimina sus registros. No usar sobre una base con datos personales.
const { chromium } = require('playwright');
const { expect } = require('playwright/test');
const fs = require('node:fs');
const output = process.env.TEST_OUTPUT || require('node:path').join(require('node:os').tmpdir(), 'bookfly-tp4-verificacion');
const baseURL = process.env.BASE_URL || 'http://localhost:8080';
fs.mkdirSync(output, {recursive:true});
(async () => {
 const browser = await chromium.launch({headless:true, channel:"chrome"});
 const page = await browser.newPage({viewport:{width:1440,height:1080}});
 const errors=[]; const mutations=[]; let navigations=0;
 page.on('pageerror',e=>errors.push(e.message));
 page.on('framenavigated', frame=>{if(frame===page.mainFrame())navigations++;});
 page.on('request',r=>{if(['POST','DELETE'].includes(r.method()))mutations.push({method:r.method(),url:r.url(),data:r.postDataJSON()});});
 const submit = async (entity, values) => {
   const form=page.locator(`#${entity}-form`);
   for(const [name,value] of Object.entries(values)){
     const input=form.locator(`[name="${name}"]`);
     if(await input.evaluate(e=>e.tagName==='SELECT'))await input.selectOption({label:value});
     else await input.fill(value);
   }
   await form.locator('button').click();
   await expect(page.locator(`#${entity}-message`)).toContainText('creado.');
   await expect(form.locator('button')).toBeEnabled();
 };
 const remove = async (entity,name) => {
   await page.locator(`#${entity}-list`).getByRole('button',{name:`Eliminar ${name}`,exact:true}).click();
   await expect(page.locator(`#${entity}-message`)).toContainText('eliminado.');
   await expect(page.locator(`#${entity}-form button`)).toBeEnabled();
 };
 try {
   await page.goto(baseURL);
   for(const entity of ['libros','autores','editoriales','usuarios'])await expect(page.locator(`#${entity}-count`)).toHaveText('0 registros');
   console.log('PASS: carga inicial y cuatro listas vacías');
   await submit('autores',{nombre:'Ursula Test',pais:'Estados Unidos',fecha_nacimiento:'1929-10-21',biografia:'Autora de prueba.'});
   const injection='<img src=x onerror="window.inyectado=true"> Editorial';
   await submit('editoriales',{nombre:injection,sede:'Buenos Aires'});
   await expect(page.locator('#editoriales-list img')).toHaveCount(0);
   await expect(page.locator('#editoriales-list')).toContainText(injection);
   if(await page.evaluate(()=>window.inyectado))throw new Error('XSS ejecutado');
   await submit('usuarios',{nombre:'lector_tp4',email:'lector_tp4@example.com',contrasena:'secreta123'});
   await expect(page.locator('#usuario-contrasena')).toHaveValue('');
   await expect(page.locator('#usuarios-list')).not.toContainText('secreta123');
   await submit('libros',{titulo:'Un libro de prueba',id_autor:'Ursula Test',id_editorial:injection,fecha_lanzamiento:'2020-01-15',cantidad_paginas:'250'});
   await expect(page.locator('#libros-list')).toContainText('Ursula Test');
   if(navigations!==1)throw new Error('Hubo recarga al crear');
   const bookPost=mutations.find(r=>r.method==='POST'&&r.url.endsWith('/libros'));
   if(typeof bookPost.data.id_autor!=='number'||bookPost.data.cantidad_paginas!==250)throw new Error('Tipos JSON incorrectos');
   console.log('PASS: cuatro altas, relaciones, JSON numérico, sin recargas y texto seguro');
   await page.addStyleTag({content:'html { scroll-behavior: auto !important; }'});
   await page.evaluate(() => window.scrollTo(0, 0));
   await page.screenshot({path:`${output}/escritorio.png`,fullPage:true});
   await page.setViewportSize({width:390,height:844});
   await page.evaluate(() => window.scrollTo(0, 0));
   await page.screenshot({path:`${output}/celular.png`,fullPage:true});
   if(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth))throw new Error('Desborde horizontal móvil');
   console.log('PASS: diseño móvil sin desborde');
   await page.reload();
   await expect(page.locator('#libros-list')).toContainText('Un libro de prueba');
   await expect(page.locator('#libros-form button')).toBeEnabled();
   console.log('PASS: persistencia después de recargar');
   await page.locator('#autores-list').getByRole('button',{name:'Eliminar Ursula Test',exact:true}).click();
   await expect(page.locator('#autores-message')).toContainText('datos relacionados');
   await expect(page.locator('#autores-list')).toContainText('Ursula Test');
   console.log('PASS: conflicto de clave foránea visible y autor conservado');
   await page.locator('#usuario-nombre').fill('lector_tp4');
   await page.locator('#usuario-email').fill('lector_tp4@example.com');
   await page.locator('#usuario-contrasena').fill('otra123');
   await page.locator('#usuarios-form button').click();
   await expect(page.locator('#usuarios-message')).toContainText('registrado');
   await expect(page.locator('#usuario-nombre')).toHaveValue('lector_tp4');
   await expect(page.locator('#usuarios-count')).toHaveText('1 registro');
   console.log('PASS: duplicados muestran error y conservan formulario');
   await page.locator('#editorial-nombre').fill('   ');
   await page.locator('#editorial-sede').fill('   ');
   const before=mutations.length;
   await page.locator('#editoriales-form button').click();
   await expect(page.locator('#editoriales-message')).toContainText('no solo espacios');
   if(mutations.length!==before)throw new Error('Se enviaron datos vacíos');
   console.log('PASS: campos con espacios rechazados');
   await remove('libros','Un libro de prueba');
   await remove('autores','Ursula Test');
   await remove('editoriales',injection);
   await remove('usuarios','lector_tp4');
   await submit('editoriales',{nombre:'Segunda editorial',sede:'Tandil'});
   await remove('editoriales','Segunda editorial');
   for(const entity of ['libros','autores','editoriales','usuarios'])await expect(page.locator(`#${entity}-count`)).toHaveText('0 registros');
   console.log('PASS: borrado de las cuatro entidades y delegación tras regenerar botones');
   await page.route('**/libros',route=>route.request().method()==='GET'?route.abort():route.continue());
   await page.reload();
   await expect(page.locator('#libros-message')).toContainText('No se pudo actualizar');
   await expect(page.locator('#libros-count')).toHaveText('Sin actualizar');
   await expect(page.locator('#libros-list .empty')).toHaveCount(0);
   await page.unroute('**/libros');
   console.log('PASS: error de red distinguido de lista vacía');
   await page.reload();
   await expect(page.locator('#editoriales-form button')).toBeEnabled();
   let failRefresh=false;
   await page.route('**/editoriales',async route=>{
     if(route.request().method()==='POST'){await route.continue();failRefresh=true;}
     else if(failRefresh)await route.abort();
     else await route.continue();
   });
   await submit('editoriales',{nombre:'Guardada con fallo de lectura',sede:'Tandil'});
   await expect(page.locator('#editoriales-message')).toContainText('El cambio se guardó');
   await expect(page.locator('#editorial-nombre')).toHaveValue('');
   await page.unroute('**/editoriales');
   await page.reload();
   await expect(page.locator('#editoriales-list')).toContainText('Guardada con fallo de lectura');
   await remove('editoriales','Guardada con fallo de lectura');
   console.log('PASS: alta confirmada no se confunde con fallo posterior de lectura');
   if(errors.length)throw new Error(errors.join('\n'));
   fs.writeFileSync(`${output}/resultado.json`,JSON.stringify({status:'passed',browser:'Google Chrome',mutations:mutations.length,consoleErrors:errors},null,2));
   console.log('PASS: sin errores JavaScript');
 } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exit(1);});
