import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';
import { SbbHeaderLeanModule } from '@sbb-esta/angular/header-lean';
import { SbbIconModule } from '@sbb-esta/angular/icon';
import { environment } from '../environments/environment';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet, SbbHeaderLeanModule, SbbIconModule],
  template: `
    <sbb-header-lean label="ngssc App"></sbb-header-lean>
    <div class="content">
      <h1>
        Welcome to {{title}} <sbb-icon svgIcon="cloud-small"></sbb-icon>!
      </h1>
      <div>{{ title }} app is running!</div>
    </div>
    <h2>Here are some links to help you start: </h2>
    <ul>
      <li>
        <h2><a target="_blank" rel="noopener" href="https://angular.io/tutorial">Tour of Heroes</a></h2>
      </li>
      <li>
        <h2><a target="_blank" rel="noopener" href="https://angular.io/cli">CLI Documentation</a></h2>
      </li>
      <li>
        <h2><a target="_blank" rel="noopener" href="https://blog.angular.io/">Angular blog</a></h2>
      </li>
    </ul>
    <router-outlet></router-outlet>
  `,
  styles: [],
})
export class AppComponent {
  title = environment.label;
}
